package longtailstorelib

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/DanEngelbrecht/golongtail/longtaillib"
	"github.com/pkg/errors"
)

// AccessType defines how we will access the data in the store
type AccessType int

const (
	// Init - read/write access with forced rebuild of store index
	Init AccessType = iota
	// ReadWrite - read/write access with optional rebuild of store index
	ReadWrite
	// ReadOnly - read only access
	ReadOnly
)

type putBlockMessage struct {
	storedBlock      longtaillib.Longtail_StoredBlock
	asyncCompleteAPI longtaillib.Longtail_AsyncPutStoredBlockAPI
}

type getBlockMessage struct {
	blockHash        uint64
	asyncCompleteAPI longtaillib.Longtail_AsyncGetStoredBlockAPI
}

type prefetchBlockMessage struct {
	blockHash uint64
}

type preflightGetMessage struct {
	chunkHashes []uint64
}

type blockIndexMessage struct {
	blockIndex longtaillib.Longtail_BlockIndex
}

type getExistingContentMessage struct {
	chunkHashes          []uint64
	minBlockUsagePercent uint32
	asyncCompleteAPI     longtaillib.Longtail_AsyncGetExistingContentAPI
}

type pendingPrefetchedBlock struct {
	storedBlock       longtaillib.Longtail_StoredBlock
	completeCallbacks []longtaillib.Longtail_AsyncGetStoredBlockAPI
}

type remoteStore struct {
	jobAPI        longtaillib.Longtail_JobAPI
	blobStore     BlobStore
	defaultClient BlobClient

	workerCount int

	putBlockChan           chan putBlockMessage
	getBlockChan           chan getBlockMessage
	preflightGetChan       chan preflightGetMessage
	prefetchBlockChan      chan prefetchBlockMessage
	blockIndexChan         chan blockIndexMessage
	getExistingContentChan chan getExistingContentMessage
	workerFlushChan        chan int
	workerFlushReplyChan   chan int
	indexFlushChan         chan int
	indexFlushReplyChan    chan int
	workerErrorChan        chan error
	prefetchMemory         int64
	maxPrefetchMemory      int64

	fetchedBlocksSync sync.Mutex
	fetchedBlocks     map[uint64]bool
	prefetchBlocks    map[uint64]*pendingPrefetchedBlock

	stats longtaillib.BlockStoreStats
}

// String() ...
func (s *remoteStore) String() string {
	return s.defaultClient.String()
}

func putStoredBlock(
	s *remoteStore,
	blobClient BlobClient,
	blockIndexMessages chan<- blockIndexMessage,
	storedBlock longtaillib.Longtail_StoredBlock) error {

	atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_PutStoredBlock_Count], 1)

	blockIndex := storedBlock.GetBlockIndex()
	blockHash := blockIndex.GetBlockHash()
	key := GetBlockPath("chunks", blockHash)
	blob, errno := longtaillib.WriteStoredBlockToBuffer(storedBlock)
	if errno != 0 {
		return longtaillib.ErrnoToError(errno, longtaillib.ErrEIO)
	}

	retryCount, err := writeBlobWithRetry(blobClient, key, false, blob)
	atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_PutStoredBlock_RetryCount], uint64(retryCount))
	if err != nil {
		atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_PutStoredBlock_FailCount], 1)
		return err
	}

	atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_PutStoredBlock_Byte_Count], (uint64)(len(blob)))
	atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_PutStoredBlock_Chunk_Count], (uint64)(blockIndex.GetChunkCount()))

	blockIndexCopy, err := blockIndex.Copy()
	if err != nil {
		return err
	}
	blockIndexMessages <- blockIndexMessage{blockIndex: blockIndexCopy}
	return nil
}

func getStoredBlock(
	s *remoteStore,
	blobClient BlobClient,
	blockHash uint64) (longtaillib.Longtail_StoredBlock, error) {

	atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_GetStoredBlock_Count], 1)

	key := GetBlockPath("chunks", blockHash)

	storedBlockData, retryCount, err := readBlobWithRetry(blobClient, key)
	atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_GetStoredBlock_RetryCount], uint64(retryCount))

	if err != nil || storedBlockData == nil {
		atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_GetStoredBlock_FailCount], 1)
		return longtaillib.Longtail_StoredBlock{}, err
	}

	storedBlock, errno := longtaillib.ReadStoredBlockFromBuffer(storedBlockData)
	if errno != 0 {
		atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_GetStoredBlock_FailCount], 1)
		return longtaillib.Longtail_StoredBlock{}, longtaillib.ErrnoToError(errno, longtaillib.ErrEIO)
	}

	atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_GetStoredBlock_Byte_Count], (uint64)(len(storedBlockData)))
	blockIndex := storedBlock.GetBlockIndex()
	if blockIndex.GetBlockHash() != blockHash {
		atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_GetStoredBlock_FailCount], 1)
		return longtaillib.Longtail_StoredBlock{}, longtaillib.ErrnoToError(longtaillib.EBADF, longtaillib.ErrEBADF)
	}
	atomic.AddUint64(&s.stats.StatU64[longtaillib.Longtail_BlockStoreAPI_StatU64_GetStoredBlock_Chunk_Count], (uint64)(blockIndex.GetChunkCount()))
	return storedBlock, nil
}

func fetchBlock(
	ctx context.Context,
	s *remoteStore,
	client BlobClient,
	getMsg getBlockMessage) {
	s.fetchedBlocksSync.Lock()
	prefetchedBlock, exists := s.prefetchBlocks[getMsg.blockHash]
	if exists {
		storedBlock := prefetchedBlock.storedBlock
		if storedBlock.IsValid() {
			delete(s.prefetchBlocks, getMsg.blockHash)
			blockSize := -int64(storedBlock.GetBlockSize())
			atomic.AddInt64(&s.prefetchMemory, blockSize)
			s.fetchedBlocksSync.Unlock()
			getMsg.asyncCompleteAPI.OnComplete(storedBlock, 0)
			return
		}
		prefetchedBlock.completeCallbacks = append(prefetchedBlock.completeCallbacks, getMsg.asyncCompleteAPI)
		s.fetchedBlocksSync.Unlock()
		return
	}
	prefetchedBlock = &pendingPrefetchedBlock{storedBlock: longtaillib.Longtail_StoredBlock{}}
	s.prefetchBlocks[getMsg.blockHash] = prefetchedBlock
	s.fetchedBlocks[getMsg.blockHash] = true
	s.fetchedBlocksSync.Unlock()
	storedBlock, getStoredBlockErr := getStoredBlock(s, client, getMsg.blockHash)
	s.fetchedBlocksSync.Lock()
	completeCallbacks := prefetchedBlock.completeCallbacks
	delete(s.prefetchBlocks, getMsg.blockHash)
	s.fetchedBlocksSync.Unlock()
	for _, c := range completeCallbacks {
		if getStoredBlockErr != nil {
			c.OnComplete(longtaillib.Longtail_StoredBlock{}, longtaillib.ErrorToErrno(getStoredBlockErr, longtaillib.EIO))
			continue
		}
		buf, errno := longtaillib.WriteStoredBlockToBuffer(storedBlock)
		if errno != 0 {
			c.OnComplete(longtaillib.Longtail_StoredBlock{}, errno)
			continue
		}
		blockCopy, errno := longtaillib.ReadStoredBlockFromBuffer(buf)
		if errno != 0 {
			c.OnComplete(longtaillib.Longtail_StoredBlock{}, errno)
			continue
		}
		c.OnComplete(blockCopy, 0)
	}
	getMsg.asyncCompleteAPI.OnComplete(storedBlock, longtaillib.ErrorToErrno(getStoredBlockErr, longtaillib.EIO))
}

func prefetchBlock(
	ctx context.Context,
	s *remoteStore,
	client BlobClient,
	prefetchMsg prefetchBlockMessage) {
	s.fetchedBlocksSync.Lock()
	_, exists := s.fetchedBlocks[prefetchMsg.blockHash]
	if exists {
		// Already fetched
		s.fetchedBlocksSync.Unlock()
		return
	}
	s.fetchedBlocks[prefetchMsg.blockHash] = true
	prefetchedBlock := &pendingPrefetchedBlock{storedBlock: longtaillib.Longtail_StoredBlock{}}
	s.prefetchBlocks[prefetchMsg.blockHash] = prefetchedBlock
	s.fetchedBlocksSync.Unlock()

	storedBlock, getErr := getStoredBlock(s, client, prefetchMsg.blockHash)
	if getErr != nil {
		return
	}

	s.fetchedBlocksSync.Lock()
	prefetchedBlock = s.prefetchBlocks[prefetchMsg.blockHash]
	completeCallbacks := prefetchedBlock.completeCallbacks
	if len(completeCallbacks) == 0 {
		// Nobody is actively waiting for the block
		blockSize := int64(storedBlock.GetBlockSize())
		prefetchedBlock.storedBlock = storedBlock
		atomic.AddInt64(&s.prefetchMemory, blockSize)
		s.fetchedBlocksSync.Unlock()
		return
	}
	delete(s.prefetchBlocks, prefetchMsg.blockHash)
	s.fetchedBlocksSync.Unlock()
	for i := 1; i < len(completeCallbacks)-1; i++ {
		c := completeCallbacks[i]
		if getErr != nil {
			c.OnComplete(longtaillib.Longtail_StoredBlock{}, longtaillib.ErrorToErrno(getErr, longtaillib.EIO))
			continue
		}
		buf, errno := longtaillib.WriteStoredBlockToBuffer(storedBlock)
		if errno != 0 {
			c.OnComplete(longtaillib.Longtail_StoredBlock{}, errno)
			continue
		}
		blockCopy, errno := longtaillib.ReadStoredBlockFromBuffer(buf)
		if errno != 0 {
			c.OnComplete(longtaillib.Longtail_StoredBlock{}, errno)
			continue
		}
		c.OnComplete(blockCopy, 0)
	}
	completeCallbacks[0].OnComplete(storedBlock, longtaillib.ErrorToErrno(getErr, longtaillib.EIO))
}

func remoteWorker(
	ctx context.Context,
	s *remoteStore,
	putBlockMessages <-chan putBlockMessage,
	getBlockMessages <-chan getBlockMessage,
	prefetchBlockChan <-chan prefetchBlockMessage,
	blockIndexMessages chan<- blockIndexMessage,
	flushMessages <-chan int,
	flushReplyMessages chan<- int,
	accessType AccessType) error {
	client, err := s.blobStore.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, s.blobStore.String())
	}
	defer client.Close()
	run := true
	for run {
		received := 0
		select {
		case putMsg, more := <-putBlockMessages:
			if more {
				received++
				if accessType == ReadOnly {
					putMsg.asyncCompleteAPI.OnComplete(longtaillib.EACCES)
					continue
				}
				err := putStoredBlock(s, client, blockIndexMessages, putMsg.storedBlock)
				putMsg.asyncCompleteAPI.OnComplete(longtaillib.ErrorToErrno(err, longtaillib.EIO))
			} else {
				run = false
			}
		case getMsg := <-getBlockMessages:
			received++
			fetchBlock(ctx, s, client, getMsg)
		default:
		}
		if received == 0 {
			if s.prefetchMemory < s.maxPrefetchMemory {
				select {
				case _ = <-flushMessages:
					flushReplyMessages <- 0
				case putMsg, more := <-putBlockMessages:
					if more {
						if accessType == ReadOnly {
							putMsg.asyncCompleteAPI.OnComplete(longtaillib.EACCES)
							continue
						}
						err := putStoredBlock(s, client, blockIndexMessages, putMsg.storedBlock)
						putMsg.asyncCompleteAPI.OnComplete(longtaillib.ErrorToErrno(err, longtaillib.EIO))
					} else {
						run = false
					}
				case getMsg := <-getBlockMessages:
					fetchBlock(ctx, s, client, getMsg)
				case prefetchMsg := <-prefetchBlockChan:
					prefetchBlock(ctx, s, client, prefetchMsg)
				}
			} else {
				select {
				case _ = <-flushMessages:
					flushReplyMessages <- 0
				case putMsg, more := <-putBlockMessages:
					if more {
						if accessType == ReadOnly {
							putMsg.asyncCompleteAPI.OnComplete(longtaillib.EACCES)
							continue
						}
						err := putStoredBlock(s, client, blockIndexMessages, putMsg.storedBlock)
						putMsg.asyncCompleteAPI.OnComplete(longtaillib.ErrorToErrno(err, longtaillib.EIO))
					} else {
						run = false
					}
				case getMsg := <-getBlockMessages:
					fetchBlock(ctx, s, client, getMsg)
				}
			}
		}
	}
	return nil
}

func buildStoreIndexFromStoreBlocks(
	ctx context.Context,
	s *remoteStore,
	blobClient BlobClient) (longtaillib.Longtail_StoreIndex, error) {

	var items []string
	blobs, err := blobClient.GetObjects("chunks")
	if err != nil {
		return longtaillib.Longtail_StoreIndex{}, err
	}

	for _, blob := range blobs {
		if blob.Size == 0 {
			continue
		}
		if strings.HasSuffix(blob.Name, ".lsb") {
			items = append(items, blob.Name)
		}
	}

	return getStoreIndexFromBlocks(ctx, s, blobClient, items)
}

func storeIndexWorkerReplyErrorState(
	blockIndexMessages <-chan blockIndexMessage,
	getExistingContentMessages <-chan getExistingContentMessage,
	flushMessages <-chan int,
	flushReplyMessages chan<- int) {
	for {
		select {
		case _ = <-flushMessages:
			flushReplyMessages <- 0
		case _, more := <-blockIndexMessages:
			if !more {
				return
			}
		case getExistingContentMessage := <-getExistingContentMessages:
			getExistingContentMessage.asyncCompleteAPI.OnComplete(longtaillib.Longtail_StoreIndex{}, longtaillib.EINVAL)
		}
	}
}

func onPreflighMessage(
	s *remoteStore,
	storeIndex longtaillib.Longtail_StoreIndex,
	message preflightGetMessage,
	prefetchBlockMessages chan<- prefetchBlockMessage) {
	existingStoreIndex, errno := longtaillib.GetExistingStoreIndex(storeIndex, message.chunkHashes, 0)
	if errno != 0 {
		log.Printf("WARNING: onPreflighMessage longtaillib.GetExistingStoreIndex() failed with %v", longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM))
		return
	}
	defer existingStoreIndex.Dispose()
	blockHashes := existingStoreIndex.GetBlockHashes()
	for _, blockHash := range blockHashes {
		prefetchBlockMessages <- prefetchBlockMessage{blockHash: blockHash}
	}
}

func onGetExistingContentMessage(
	s *remoteStore,
	storeIndex longtaillib.Longtail_StoreIndex,
	message getExistingContentMessage) {
	existingStoreIndex, errno := longtaillib.GetExistingStoreIndex(storeIndex, message.chunkHashes, message.minBlockUsagePercent)
	if errno != 0 {
		message.asyncCompleteAPI.OnComplete(longtaillib.Longtail_StoreIndex{}, errno)
		return
	}
	message.asyncCompleteAPI.OnComplete(existingStoreIndex, 0)
}

func contentIndexWorker(
	ctx context.Context,
	s *remoteStore,
	preflightGetMessages <-chan preflightGetMessage,
	prefetchBlockMessages chan<- prefetchBlockMessage,
	blockIndexMessages <-chan blockIndexMessage,
	getExistingContentMessages <-chan getExistingContentMessage,
	flushMessages <-chan int,
	flushReplyMessages chan<- int,
	accessType AccessType) error {

	client, err := s.blobStore.NewClient(ctx)
	if err != nil {
		storeIndexWorkerReplyErrorState(blockIndexMessages, getExistingContentMessages, flushMessages, flushReplyMessages)
		return errors.Wrap(err, s.blobStore.String())
	}
	defer client.Close()

	var errno int

	saveStoreIndex := false

	var storeIndex longtaillib.Longtail_StoreIndex
	if accessType == Init {
		saveStoreIndex = true
	} else {
		storeIndex, err = readStoreIndex(client)
		if err != nil {
			log.Printf("contentIndexWorker: readStoreStoreIndex() failed with %v", err)
		}
	}
	defer storeIndex.Dispose()

	if !storeIndex.IsValid() {
		if accessType == ReadOnly {
			storeIndex, errno = longtaillib.CreateStoreIndexFromBlocks([]longtaillib.Longtail_BlockIndex{})
			if errno != 0 {
				storeIndexWorkerReplyErrorState(blockIndexMessages, getExistingContentMessages, flushMessages, flushReplyMessages)
				return errors.Wrapf(longtaillib.ErrnoToError(longtaillib.EACCES, longtaillib.ErrEACCES), "contentIndexWorker: CreateStoreIndexFromBlocks() failed")
			}
		} else {
			storeIndex, err = buildStoreIndexFromStoreBlocks(
				ctx,
				s,
				client)

			if err != nil {
				storeIndexWorkerReplyErrorState(blockIndexMessages, getExistingContentMessages, flushMessages, flushReplyMessages)
				return errors.Wrapf(longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM), "contentIndexWorker: buildStoreIndexFromStoreBlocks() failed")
			}
			log.Printf("Rebuilt remote index with %d blocks\n", len(storeIndex.GetBlockHashes()))
			err := writeStoreIndex(client, storeIndex)
			if err != nil {
				log.Printf("Failed to update store index in store %s\n", s.String())
				saveStoreIndex = true
			}
		}
	}

	var addedBlockIndexes []longtaillib.Longtail_BlockIndex
	defer func(addedBlockIndexes []longtaillib.Longtail_BlockIndex) {
		for _, blockIndex := range addedBlockIndexes {
			blockIndex.Dispose()
		}
	}(addedBlockIndexes)

	run := true
	for run {
		received := 0
		select {
		case preflightGetMsg := <-preflightGetMessages:
			received++
			if len(addedBlockIndexes) > 0 {
				updatedStoreIndex, err := updateStoreIndex(storeIndex, addedBlockIndexes)
				if err != nil {
					log.Printf("WARNING: Failed to update store index with added blocks %v", err)
					continue
				}
				storeIndex.Dispose()
				storeIndex = updatedStoreIndex
			}
			onPreflighMessage(s, storeIndex, preflightGetMsg, prefetchBlockMessages)
		case blockIndexMsg, more := <-blockIndexMessages:
			if more {
				received++
				addedBlockIndexes = append(addedBlockIndexes, blockIndexMsg.blockIndex)
			} else {
				run = false
			}
		case getExistingContentMessage := <-getExistingContentMessages:
			received++
			if len(addedBlockIndexes) > 0 {
				updatedStoreIndex, err := updateStoreIndex(storeIndex, addedBlockIndexes)
				if err != nil {
					log.Printf("WARNING: Failed to update store index with added blocks %v", err)
					getExistingContentMessage.asyncCompleteAPI.OnComplete(longtaillib.Longtail_StoreIndex{}, longtaillib.ErrorToErrno(err, longtaillib.EIO))
					continue
				}
				storeIndex.Dispose()
				storeIndex = updatedStoreIndex
			}
			onGetExistingContentMessage(s, storeIndex, getExistingContentMessage)
		default:
		}

		if received == 0 {
			select {
			case _ = <-flushMessages:
				flushReplyMessages <- 0
			case preflightGetMsg := <-preflightGetMessages:
				if len(addedBlockIndexes) > 0 {
					updatedStoreIndex, err := updateStoreIndex(storeIndex, addedBlockIndexes)
					if err != nil {
						log.Printf("WARNING: Failed to update store index with added blocks %v", err)
						continue
					}
					storeIndex.Dispose()
					storeIndex = updatedStoreIndex
				}
				onPreflighMessage(s, storeIndex, preflightGetMsg, prefetchBlockMessages)
			case blockIndexMsg, more := <-blockIndexMessages:
				if more {
					addedBlockIndexes = append(addedBlockIndexes, blockIndexMsg.blockIndex)
				} else {
					run = false
				}
			case getExistingContentMessage := <-getExistingContentMessages:
				if len(addedBlockIndexes) > 0 {
					updatedStoreIndex, err := updateStoreIndex(storeIndex, addedBlockIndexes)
					if err != nil {
						log.Printf("WARNING: Failed to update store index with added blocks %v", err)
						getExistingContentMessage.asyncCompleteAPI.OnComplete(longtaillib.Longtail_StoreIndex{}, longtaillib.ErrorToErrno(err, longtaillib.EIO))
						continue
					}
					storeIndex.Dispose()
					storeIndex = updatedStoreIndex
				}
				onGetExistingContentMessage(s, storeIndex, getExistingContentMessage)
			}
		}
	}

	if accessType == ReadOnly {
		return nil
	}

	if len(addedBlockIndexes) > 0 {
		saveStoreIndex = true
	}

	if saveStoreIndex {
		addedStoreIndex, errno := longtaillib.CreateStoreIndexFromBlocks(addedBlockIndexes)
		defer addedStoreIndex.Dispose()
		if errno != 0 {
			return errors.Wrap(longtaillib.ErrnoToError(errno, longtaillib.ErrENOMEM), s.blobStore.String())
		}
		err := writeStoreIndex(client, addedStoreIndex)
		if err != nil {
			storeIndexWorkerReplyErrorState(blockIndexMessages, getExistingContentMessages, flushMessages, flushReplyMessages)
		}
	}
	return nil
}

// NewRemoteBlockStore ...
func NewRemoteBlockStore(
	jobAPI longtaillib.Longtail_JobAPI,
	blobStore BlobStore,
	accessType AccessType) (longtaillib.BlockStoreAPI, error) {
	ctx := context.Background()
	defaultClient, err := blobStore.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, blobStore.String())
	}

	s := &remoteStore{
		jobAPI:        jobAPI,
		blobStore:     blobStore,
		defaultClient: defaultClient}

	s.workerCount = runtime.NumCPU()
	s.putBlockChan = make(chan putBlockMessage, s.workerCount*8)
	s.getBlockChan = make(chan getBlockMessage, s.workerCount*2048)
	s.prefetchBlockChan = make(chan prefetchBlockMessage, s.workerCount*2048)
	s.preflightGetChan = make(chan preflightGetMessage, 16)
	s.blockIndexChan = make(chan blockIndexMessage, s.workerCount*2048)
	s.getExistingContentChan = make(chan getExistingContentMessage, 16)
	s.workerFlushChan = make(chan int, s.workerCount)
	s.workerFlushReplyChan = make(chan int, s.workerCount)
	s.indexFlushChan = make(chan int, 1)
	s.indexFlushReplyChan = make(chan int, 1)
	s.workerErrorChan = make(chan error, 1+s.workerCount)

	s.prefetchMemory = 0
	s.maxPrefetchMemory = 512 * 1024 * 1024

	s.fetchedBlocks = map[uint64]bool{}
	s.prefetchBlocks = map[uint64]*pendingPrefetchedBlock{}

	go func() {
		err := contentIndexWorker(ctx, s, s.preflightGetChan, s.prefetchBlockChan, s.blockIndexChan, s.getExistingContentChan, s.indexFlushChan, s.indexFlushReplyChan, accessType)
		s.workerErrorChan <- err
	}()

	for i := 0; i < s.workerCount; i++ {
		go func() {
			err := remoteWorker(ctx, s, s.putBlockChan, s.getBlockChan, s.prefetchBlockChan, s.blockIndexChan, s.workerFlushChan, s.workerFlushReplyChan, accessType)
			s.workerErrorChan <- err
		}()
	}

	return s, nil
}

// GetBlockPath ...
func GetBlockPath(basePath string, blockHash uint64) string {
	fileName := fmt.Sprintf("0x%016x.lsb", blockHash)
	dir := filepath.Join(basePath, fileName[2:6])
	name := filepath.Join(dir, fileName)
	name = strings.Replace(name, "\\", "/", -1)
	return name
}

// PutStoredBlock ...
func (s *remoteStore) PutStoredBlock(storedBlock longtaillib.Longtail_StoredBlock, asyncCompleteAPI longtaillib.Longtail_AsyncPutStoredBlockAPI) int {
	s.putBlockChan <- putBlockMessage{storedBlock: storedBlock, asyncCompleteAPI: asyncCompleteAPI}
	return 0
}

// PreflightGet ...
func (s *remoteStore) PreflightGet(chunkHashes []uint64) int {
	s.preflightGetChan <- preflightGetMessage{chunkHashes: chunkHashes}
	return 0
}

// GetStoredBlock ...
func (s *remoteStore) GetStoredBlock(blockHash uint64, asyncCompleteAPI longtaillib.Longtail_AsyncGetStoredBlockAPI) int {
	s.getBlockChan <- getBlockMessage{blockHash: blockHash, asyncCompleteAPI: asyncCompleteAPI}
	return 0
}

// GetExistingContent ...
func (s *remoteStore) GetExistingContent(
	chunkHashes []uint64,
	minBlockUsagePercent uint32,
	asyncCompleteAPI longtaillib.Longtail_AsyncGetExistingContentAPI) int {
	s.getExistingContentChan <- getExistingContentMessage{chunkHashes: chunkHashes, minBlockUsagePercent: minBlockUsagePercent, asyncCompleteAPI: asyncCompleteAPI}
	return 0
}

// GetStats ...
func (s *remoteStore) GetStats() (longtaillib.BlockStoreStats, int) {
	return s.stats, 0
}

// Flush ...
func (s *remoteStore) Flush(asyncCompleteAPI longtaillib.Longtail_AsyncFlushAPI) int {
	go func() {
		any_errno := 0
		for i := 0; i < s.workerCount; i++ {
			s.workerFlushChan <- 1
		}
		for i := 0; i < s.workerCount; i++ {
			select {
			case errno := <-s.workerFlushReplyChan:
				if errno != 0 && any_errno == 0 {
					any_errno = errno
				}
			}
		}
		s.indexFlushChan <- 1
		select {
		case errno := <-s.indexFlushReplyChan:
			if errno != 0 && any_errno == 0 {
				any_errno = errno
			}
		}
		asyncCompleteAPI.OnComplete(any_errno)
	}()
	return 0
}

// Close ...
func (s *remoteStore) Close() {
	close(s.putBlockChan)
	for i := 0; i < s.workerCount; i++ {
		select {
		case err := <-s.workerErrorChan:
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	close(s.blockIndexChan)
	select {
	case err := <-s.workerErrorChan:
		if err != nil {
			log.Fatal(err)
		}
	}
	s.defaultClient.Close()
}
