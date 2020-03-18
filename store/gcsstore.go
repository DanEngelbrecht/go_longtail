package store

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/DanEngelbrecht/golongtail/lib"
	"github.com/pkg/errors"
)

type gcsFileStorage struct {
}

func (fileStorage *gcsFileStorage) ReadFromPath(ctx context.Context, path string) ([]byte, error) {
	u, err := url.Parse(path)
	if u.Scheme != "gs" {
		return nil, fmt.Errorf("invalid scheme '%s', expected 'gs'", u.Scheme)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	bucketName := u.Host
	bucket := client.Bucket(bucketName)

	objHandle := bucket.Object(u.Path[1:])
	objReader, err := objHandle.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer objReader.Close()

	return ioutil.ReadAll(objReader)
}

func (fileStorage *gcsFileStorage) WriteToPath(ctx context.Context, path string, data []byte) error {
	u, err := url.Parse(path)
	if u.Scheme != "gs" {
		return fmt.Errorf("invalid scheme '%s', expected 'gs'", u.Scheme)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	bucketName := u.Host
	bucket := client.Bucket(bucketName)

	objHandle := bucket.Object(u.Path[1:])
	{
		objWriter := objHandle.NewWriter(ctx)
		_, err := objWriter.Write(data)
		objWriter.Close()
		if err != nil {
			return err
		}
	}
	_, err = objHandle.Update(ctx, storage.ObjectAttrsToUpdate{ContentType: "application/octet-stream"})
	return err
}

func (fileStorage *gcsFileStorage) Close() {
}

// NewGCSFileStorage ...
func NewGCSFileStorage(u *url.URL) (FileStorage, error) {
	if u.Scheme != "gs" {
		return nil, fmt.Errorf("invalid scheme '%s', expected 'gs'", u.Scheme)
	}
	s := &gcsFileStorage{}
	return s, nil
}

type putBlockMessage struct {
	storedBlock      lib.Longtail_StoredBlock
	asyncCompleteAPI lib.Longtail_AsyncCompleteAPI
}

type getBlockMessage struct {
	blockHash        uint64
	outStoredBlock   lib.Longtail_StoredBlockPtr
	asyncCompleteAPI lib.Longtail_AsyncCompleteAPI
}

type contentIndexMessage struct {
	contentIndex lib.Longtail_ContentIndex
}

type queryContentIndexMessage struct {
}

type responseContentIndexMessage struct {
	contentIndex lib.Longtail_ContentIndex
	errno        int
}

type stopMessage struct {
}

type gcsBlockStore struct {
	url           *url.URL
	Location      string
	defaultClient *storage.Client
	defaultBucket *storage.BucketHandle

	defaultHashAPI lib.Longtail_HashAPI
	workerCount    int

	putBlockChan             chan putBlockMessage
	getBlockChan             chan getBlockMessage
	contentIndexChan         chan contentIndexMessage
	queryContentIndexChan    chan queryContentIndexMessage
	responseContentIndexChan chan responseContentIndexMessage
	stopChan                 chan stopMessage

	workerWaitGroup sync.WaitGroup
}

// String() ...
func (s *gcsBlockStore) String() string {
	return s.Location
}

func putStoredBlock(
	ctx context.Context,
	s *gcsBlockStore,
	bucket *storage.BucketHandle,
	contentIndexMessages chan<- contentIndexMessage,
	storedBlock lib.Longtail_StoredBlock,
	asyncCompleteAPI lib.Longtail_AsyncCompleteAPI) int {
	blockIndex := storedBlock.GetBlockIndex()
	blockHash := blockIndex.GetBlockHash()
	key := getBlockPath("chunks", blockHash)
	objHandle := bucket.Object(key)
	_, err := objHandle.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		blockIndexBytes, err := lib.WriteBlockIndexToBuffer(storedBlock.GetBlockIndex())
		if err != nil {
			return asyncCompleteAPI.OnComplete(lib.ENOMEM)
		}

		blockData := storedBlock.GetChunksBlockData()
		blob := append(blockIndexBytes, blockData...)

		objWriter := objHandle.NewWriter(ctx)
		_, err = objWriter.Write(blob)
		if err != nil {
			objWriter.Close()
			//		return errors.Wrap(err, s.String()+"/"+key)
			return asyncCompleteAPI.OnComplete(lib.EIO)
		}

		err = objWriter.Close()
		if err != nil {
			//		return errors.Wrap(err, s.String()+"/"+key)
			return asyncCompleteAPI.OnComplete(lib.EIO)
		}

		_, err = objHandle.Update(ctx, storage.ObjectAttrsToUpdate{ContentType: "application/octet-stream"})
		if err != nil {
			return asyncCompleteAPI.OnComplete(lib.EIO)
		}
	}

	newBlocks := []lib.Longtail_BlockIndex{blockIndex}
	addedContentIndex, err := lib.CreateContentIndexFromBlocks(s.defaultHashAPI.GetIdentifier(), newBlocks)
	if err != nil {
		return asyncCompleteAPI.OnComplete(lib.ENOMEM)
	}
	contentIndexMessages <- contentIndexMessage{contentIndex: addedContentIndex}
	return asyncCompleteAPI.OnComplete(0)
}

func getStoredBlock(
	ctx context.Context,
	s *gcsBlockStore,
	bucket *storage.BucketHandle,
	blockHash uint64,
	outStoredBlock lib.Longtail_StoredBlockPtr,
	asyncCompleteAPI lib.Longtail_AsyncCompleteAPI) int {

	key := getBlockPath("chunks", blockHash)
	objHandle := bucket.Object(key)
	obj, err := objHandle.NewReader(ctx)
	if err != nil {
		return asyncCompleteAPI.OnComplete(lib.ENOMEM)
	}
	defer obj.Close()

	storedBlockData, err := ioutil.ReadAll(obj)

	if err != nil {
		return asyncCompleteAPI.OnComplete(lib.EIO)
	}

	storedBlock, err := lib.InitStoredBlockFromData(storedBlockData)
	if err != nil {
		return asyncCompleteAPI.OnComplete(lib.ENOMEM)
	}
	outStoredBlock.Set(storedBlock)
	return asyncCompleteAPI.OnComplete(0)
}

func gcsWorker(
	ctx context.Context,
	s *gcsBlockStore,
	u *url.URL,
	putBlockMessages <-chan putBlockMessage,
	getBlockMessages <-chan getBlockMessage,
	contentIndexMessages chan<- contentIndexMessage,
	stopMessages <-chan stopMessage) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, u.String())
	}
	bucketName := u.Host
	bucket := client.Bucket(bucketName)

	for true {
		select {
		case putMsg := <-putBlockMessages:
			errno := putStoredBlock(ctx, s, bucket, contentIndexMessages, putMsg.storedBlock, putMsg.asyncCompleteAPI)
			if errno != 0 {
				log.Printf("WARNING: putStoredBlock returned: %d", errno)
			}
		case getMsg := <-getBlockMessages:
			errno := getStoredBlock(ctx, s, bucket, getMsg.blockHash, getMsg.outStoredBlock, getMsg.asyncCompleteAPI)
			if errno != 0 {
				log.Printf("WARNING: getStoredBlock returned: %d", errno)
			}
		case _ = <-stopMessages:
			s.workerWaitGroup.Done()
			return nil
		}
	}
	s.workerWaitGroup.Done()
	return nil
}

func updateRemoteContentIndex(
	ctx context.Context,
	bucket *storage.BucketHandle,
	addedContentIndex lib.Longtail_ContentIndex) error {
	storeBlob, err := lib.WriteContentIndexToBuffer(addedContentIndex)
	if err != nil {
		return err
	}
	objHandle := bucket.Object("store.lci")
	for {
		writeCondition := storage.Conditions{DoesNotExist: true}
		objAttrs, _ := objHandle.Attrs(ctx)
		if objAttrs != nil {
			writeCondition = storage.Conditions{GenerationMatch: objAttrs.Generation}
			reader, err := objHandle.If(writeCondition).NewReader(ctx)
			if err != nil {
				return err
			}
			if reader == nil {
				continue
			}
			blob, err := ioutil.ReadAll(reader)
			reader.Close()
			if err != nil {
				return err
			}

			remoteContentIndex, err := lib.ReadContentIndexFromBuffer(blob)
			if err != nil {
				return err
			}
			defer remoteContentIndex.Dispose()
			mergedContentIndex, err := lib.MergeContentIndex(remoteContentIndex, addedContentIndex)
			if err != nil {
				return err
			}
			defer mergedContentIndex.Dispose()

			storeBlob, err = lib.WriteContentIndexToBuffer(mergedContentIndex)
			if err != nil {
				return err
			}
		}
		writer := objHandle.If(writeCondition).NewWriter(ctx)
		if writer == nil {
			continue
		}
		_, err = writer.Write(storeBlob)
		if err != nil {
			writer.CloseWithError(err)
			return err
		}
		writer.Close()
		_, err = objHandle.Update(ctx, storage.ObjectAttrsToUpdate{ContentType: "application/octet-stream"})
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func contentIndexWorker(
	ctx context.Context,
	s *gcsBlockStore,
	u *url.URL,
	contentIndexMessages <-chan contentIndexMessage,
	queryContentIndexMessages <-chan queryContentIndexMessage,
	responseContentIndexMessages chan<- responseContentIndexMessage,
	stopMessages <-chan stopMessage) error {

	client, err := storage.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, u.String())
	}
	bucketName := u.Host
	bucket := client.Bucket(bucketName)

	var contentIndex lib.Longtail_ContentIndex

	objHandle := bucket.Object("store.lci")
	obj, err := objHandle.NewReader(ctx)
	if err == nil {
		defer obj.Close()
		storedContentIndexData, err := ioutil.ReadAll(obj)
		if err == nil {
			contentIndex, err = lib.ReadContentIndexFromBuffer(storedContentIndexData)
		}
	}

	if err != nil {
		hashAPI := lib.CreateBlake3HashAPI()
		defer hashAPI.Dispose()
		contentIndex, err = lib.CreateContentIndex(
			s.defaultHashAPI,
			[]uint64{},
			[]uint32{},
			[]uint32{},
			32768,
			65536)
		if err != nil {
			return err
		}
	}

	var addedContentIndex lib.Longtail_ContentIndex

	defer contentIndex.Dispose()
	defer addedContentIndex.Dispose()

	for true {
		select {
		case contentIndexMsg := <-contentIndexMessages:
			newContentIndex, err := lib.MergeContentIndex(contentIndex, contentIndexMsg.contentIndex)
			if err == nil {
				contentIndex.Dispose()
				contentIndex = newContentIndex
			}
			if !addedContentIndex.IsValid() {
				addedContentIndex = contentIndexMsg.contentIndex
			} else {
				newAddedContentIndex, err := lib.MergeContentIndex(addedContentIndex, contentIndexMsg.contentIndex)
				if err == nil {
					addedContentIndex.Dispose()
					addedContentIndex = newAddedContentIndex
				}
				contentIndexMsg.contentIndex.Dispose()
			}
		case _ = <-queryContentIndexMessages:
			{
				responseContentIndexMsg := responseContentIndexMessage{errno: lib.ENOMEM}
				buf, err := lib.WriteContentIndexToBuffer(contentIndex)
				if err == nil {
					contentIndexCopy, err := lib.ReadContentIndexFromBuffer(buf)
					if err == nil {
						responseContentIndexMsg = responseContentIndexMessage{contentIndex: contentIndexCopy, errno: 0}
					}
				}
				responseContentIndexMessages <- responseContentIndexMsg
			}
		case _ = <-stopMessages:
			if addedContentIndex.IsValid() {
				err := updateRemoteContentIndex(ctx, bucket, addedContentIndex)
				if err != nil {
					log.Printf("WARNING: Failed to write store content index: %q", err)
				}
			}
			s.workerWaitGroup.Done()
			return nil
		}
	}
	s.workerWaitGroup.Done()
	return nil
}

// NewGCSBlockStore ...
func NewGCSBlockStore(u *url.URL, defaultHashAPI lib.Longtail_HashAPI) (lib.BlockStoreAPI, error) {
	if u.Scheme != "gs" {
		return nil, fmt.Errorf("invalid scheme '%s', expected 'gs'", u.Scheme)
	}

	ctx := context.Background()
	defaultClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, u.String())
	}

	bucketName := u.Host
	defaultBucket := defaultClient.Bucket(bucketName)

	//	backingStorage := lib.CreateFSStorageAPI()

	s := &gcsBlockStore{url: u, Location: u.String(), defaultClient: defaultClient, defaultBucket: defaultBucket, defaultHashAPI: defaultHashAPI}
	s.workerCount = runtime.NumCPU() * 4
	s.putBlockChan = make(chan putBlockMessage, s.workerCount*8)
	s.getBlockChan = make(chan getBlockMessage, s.workerCount*8)
	s.contentIndexChan = make(chan contentIndexMessage, s.workerCount*8)
	s.queryContentIndexChan = make(chan queryContentIndexMessage)
	s.responseContentIndexChan = make(chan responseContentIndexMessage)
	s.stopChan = make(chan stopMessage, s.workerCount)

	go contentIndexWorker(ctx, s, u, s.contentIndexChan, s.queryContentIndexChan, s.responseContentIndexChan, s.stopChan)
	s.workerWaitGroup.Add(1)
	for i := 0; i < s.workerCount; i++ {
		go gcsWorker(ctx, s, u, s.putBlockChan, s.getBlockChan, s.contentIndexChan, s.stopChan)
	}
	s.workerWaitGroup.Add(s.workerCount)

	return s, nil
}

func getBlockPath(basePath string, blockHash uint64) string {
	sID := fmt.Sprintf("%x", blockHash)
	dir := filepath.Join(basePath, sID[0:4])
	name := filepath.Join(dir, sID) + ".lsb"
	name = strings.Replace(name, "\\", "/", -1)
	return name
}

// PutStoredBlock ...
func (s *gcsBlockStore) PutStoredBlock(storedBlock lib.Longtail_StoredBlock, asyncCompleteAPI lib.Longtail_AsyncCompleteAPI) int {
	s.putBlockChan <- putBlockMessage{storedBlock: storedBlock, asyncCompleteAPI: asyncCompleteAPI}
	return 0
}

// GetStoredBlock ...
func (s *gcsBlockStore) GetStoredBlock(blockHash uint64, outStoredBlock lib.Longtail_StoredBlockPtr, asyncCompleteAPI lib.Longtail_AsyncCompleteAPI) int {
	s.getBlockChan <- getBlockMessage{blockHash: blockHash, outStoredBlock: outStoredBlock, asyncCompleteAPI: asyncCompleteAPI}
	return 0
}

// GetIndex ...
func (s *gcsBlockStore) GetIndex(defaultHashAPIIdentifier uint32, jobAPI lib.Longtail_JobAPI, progress lib.Longtail_ProgressAPI) (lib.Longtail_ContentIndex, int) {
	s.queryContentIndexChan <- queryContentIndexMessage{}
	responseContentIndexMsg := <-s.responseContentIndexChan
	return responseContentIndexMsg.contentIndex, responseContentIndexMsg.errno
}

// GetStoredBlockPath ...
func (s *gcsBlockStore) GetStoredBlockPath(blockHash uint64) (string, int) {
	return getBlockPath("chunks", blockHash), 0
}

// Close ...
func (s *gcsBlockStore) Close() {
	for i := 0; i < s.workerCount+1; i++ {
		s.stopChan <- stopMessage{}
	}
	s.workerWaitGroup.Wait()
}