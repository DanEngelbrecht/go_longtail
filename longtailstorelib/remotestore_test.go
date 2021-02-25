package longtailstorelib

import (
	"context"
	"runtime"
	"sync"
	"testing"

	"github.com/DanEngelbrecht/golongtail/longtaillib"
)

func TestCreateRemoteBlobStore(t *testing.T) {
	blobStore, _ := NewTestBlobStore("the_path")
	jobs := longtaillib.CreateBikeshedJobAPI(uint32(runtime.NumCPU()), 0)
	defer jobs.Dispose()
	remoteStore, err := NewRemoteBlockStore(
		jobs,
		blobStore,
		ReadOnly)
	if err != nil {
		t.Errorf("TestCreateRemoveBlobStore() NewRemoteBlockStore()) %v != %v", err, nil)
	}
	storeAPI := longtaillib.CreateBlockStoreAPI(remoteStore)
	defer storeAPI.Dispose()
}

type getExistingContentCompletionAPI struct {
	wg         sync.WaitGroup
	storeIndex longtaillib.Longtail_StoreIndex
	err        int
}

func (a *getExistingContentCompletionAPI) OnComplete(storeIndex longtaillib.Longtail_StoreIndex, errno int) {
	a.storeIndex = storeIndex
	a.err = errno
	a.wg.Done()
}

type putStoredBlockCompletionAPI struct {
	wg  sync.WaitGroup
	err int
}

func (a *putStoredBlockCompletionAPI) OnComplete(errno int) {
	a.err = errno
	a.wg.Done()
}

type getStoredBlockCompletionAPI struct {
	wg          sync.WaitGroup
	storedBlock longtaillib.Longtail_StoredBlock
	err         int
}

func (a *getStoredBlockCompletionAPI) OnComplete(storedBlock longtaillib.Longtail_StoredBlock, errno int) {
	a.storedBlock = storedBlock
	a.err = errno
	a.wg.Done()
}

func getExistingContent(t *testing.T, storeAPI longtaillib.Longtail_BlockStoreAPI, chunkHashes []uint64, minBlockUsagePercent uint32) (longtaillib.Longtail_StoreIndex, int) {
	g := &getExistingContentCompletionAPI{}
	g.wg.Add(1)
	errno := storeAPI.GetExistingContent(chunkHashes, minBlockUsagePercent, longtaillib.CreateAsyncGetExistingContentAPI(g))
	if errno != 0 {
		g.wg.Done()
		return longtaillib.Longtail_StoreIndex{}, errno
	}
	g.wg.Wait()
	return g.storeIndex, g.err
}

func TestEmptyGetExistingContent(t *testing.T) {
	blobStore, _ := NewTestBlobStore("the_path")
	jobs := longtaillib.CreateBikeshedJobAPI(uint32(runtime.NumCPU()), 0)
	defer jobs.Dispose()
	remoteStore, err := NewRemoteBlockStore(
		jobs,
		blobStore,
		ReadOnly)
	if err != nil {
		t.Errorf("TestCreateRemoveBlobStore() NewRemoteBlockStore()) %v != %v", err, nil)
	}
	storeAPI := longtaillib.CreateBlockStoreAPI(remoteStore)

	chunkHashes := []uint64{1, 2, 3, 4}

	existingContent, errno := getExistingContent(t, storeAPI, chunkHashes, 0)
	defer existingContent.Dispose()
	if errno != 0 {
		t.Errorf("TestEmptyGetExistingContent() getExistingContent(t, storeAPI, chunkHashes) %d != %d", errno, 0)
	}

	if !existingContent.IsValid() {
		t.Errorf("TestEmptyGetExistingContent() g.err %t != %t", existingContent.IsValid(), true)
	}

	if existingContent.GetBlockCount() != 0 {
		t.Errorf("TestEmptyGetExistingContent() existingContent.GetBlockCount() %d != %d", existingContent.GetBlockCount(), 0)
	}

	defer storeAPI.Dispose()
}

func TestPutGetStoredBlock(t *testing.T) {
	blobStore, _ := NewTestBlobStore("the_path")
	jobs := longtaillib.CreateBikeshedJobAPI(uint32(runtime.NumCPU()), 0)
	defer jobs.Dispose()
	remoteStore, err := NewRemoteBlockStore(
		jobs,
		blobStore,
		ReadWrite)
	if err != nil {
		t.Errorf("TestPutGetStoredBlock() NewRemoteBlockStore()) %v != %v", err, nil)
	}
	storeAPI := longtaillib.CreateBlockStoreAPI(remoteStore)

	block, errno := storeBlockFromSeed(t, storeAPI, 0)
	if errno != 0 {
		t.Errorf("TestPutGetStoredBlock() storeBlock(t, storeAPI, 0) %d != %d", errno, 0)
	}
	defer block.Dispose()
	blockIndex := block.GetBlockIndex()
	blockHash := blockIndex.GetBlockHash()

	storedBlockCopy, errno := fetchBlockFromStore(t, storeAPI, blockHash)
	if errno != 0 {
		t.Errorf("TestPutGetStoredBlock() fetchBlockFromStore(t, storeAPI, 0) %d != %d", errno, 0)
	}
	defer storedBlockCopy.Dispose()

	if !storedBlockCopy.IsValid() {
		t.Errorf("TestPutGetStoredBlock() g.err %t != %t", storedBlockCopy.IsValid(), true)
	}

	validateBlockFromSeed(t, 0, storedBlockCopy)

	defer storeAPI.Dispose()
}

type flushCompletionAPI struct {
	wg  sync.WaitGroup
	err int
}

func (a *flushCompletionAPI) OnComplete(err int) {
	a.err = err
	a.wg.Done()
}

func TestGetExistingContent(t *testing.T) {
	blobStore, _ := NewTestBlobStore("the_path")
	jobs := longtaillib.CreateBikeshedJobAPI(uint32(runtime.NumCPU()), 0)
	defer jobs.Dispose()
	remoteStore, err := NewRemoteBlockStore(
		jobs,
		blobStore,
		ReadWrite)
	if err != nil {
		t.Errorf("TestPutGetStoredBlock() NewRemoteBlockStore()) %v != %v", err, nil)
	}
	storeAPI := longtaillib.CreateBlockStoreAPI(remoteStore)
	defer storeAPI.Dispose()

	b1, errno := storeBlockFromSeed(t, storeAPI, 0)
	if errno != 0 {
		t.Errorf("TestPutGetStoredBlock() storeBlock(t, storeAPI, 0) %d != %d", errno, 0)
	}
	defer b1.Dispose()

	b2, errno := storeBlockFromSeed(t, storeAPI, 10)
	if errno != 0 {
		t.Errorf("TestPutGetStoredBlock() storeBlock(t, storeAPI, 0) %d != %d", errno, 0)
	}
	defer b2.Dispose()

	b3, errno := storeBlockFromSeed(t, storeAPI, 20)
	if errno != 0 {
		t.Errorf("TestPutGetStoredBlock() storeBlock(t, storeAPI, 0) %d != %d", errno, 0)
	}
	defer b3.Dispose()

	b4, errno := storeBlockFromSeed(t, storeAPI, 30)
	if errno != 0 {
		t.Errorf("TestPutGetStoredBlock() storeBlock(t, storeAPI, 0) %d != %d", errno, 0)
	}
	defer b4.Dispose()

	b5, errno := storeBlockFromSeed(t, storeAPI, 40)
	if errno != 0 {
		t.Errorf("TestPutGetStoredBlock() storeBlock(t, storeAPI, 0) %d != %d", errno, 0)
	}
	defer b5.Dispose()

	b6, errno := storeBlockFromSeed(t, storeAPI, 50)
	if errno != 0 {
		t.Errorf("TestPutGetStoredBlock() storeBlock(t, storeAPI, 0) %d != %d", errno, 0)
	}
	defer b6.Dispose()

	b1Index := b1.GetBlockIndex()
	b2Index := b2.GetBlockIndex()
	b3Index := b3.GetBlockIndex()
	b4Index := b4.GetBlockIndex()
	b5Index := b5.GetBlockIndex()
	b6Index := b6.GetBlockIndex()

	chunkHashes := []uint64{
		b1Index.GetChunkHashes()[0], b1Index.GetChunkHashes()[1],
		b2Index.GetChunkHashes()[0], b2Index.GetChunkHashes()[1],
		b3Index.GetChunkHashes()[0], b3Index.GetChunkHashes()[1],
		b4Index.GetChunkHashes()[0], b4Index.GetChunkHashes()[1],
		b5Index.GetChunkHashes()[0], b5Index.GetChunkHashes()[1],
		b6Index.GetChunkHashes()[0]}

	remoteStoreFlushComplete := &flushCompletionAPI{}
	remoteStoreFlushComplete.wg.Add(1)
	_ = remoteStore.Flush(longtaillib.CreateAsyncFlushAPI(remoteStoreFlushComplete))
	remoteStoreFlushComplete.wg.Wait()

	existingContent, errno := getExistingContent(t, storeAPI, chunkHashes, 0)
	defer existingContent.Dispose()
	if !existingContent.IsValid() {
		t.Errorf("TestGetExistingContent() g.err %t != %t", existingContent.IsValid(), true)
	}

	if existingContent.GetBlockCount() != 6 {
		t.Errorf("TestGetExistingContent() existingContent.GetBlockCount() %d != %d", existingContent.GetBlockCount(), 6)
	}

	if existingContent.GetChunkCount() != 18 {
		t.Errorf("TestGetExistingContent() existingContent.GetChunkCount() %d != %d", existingContent.GetChunkCount(), 18)
	}
}

func TestRestoreStore(t *testing.T) {
	blobStore, _ := NewTestBlobStore("the_path")
	jobs := longtaillib.CreateBikeshedJobAPI(uint32(runtime.NumCPU()), 0)
	defer jobs.Dispose()
	remoteStore, err := NewRemoteBlockStore(
		jobs,
		blobStore,
		ReadWrite)
	if err != nil {
		t.Errorf("TestPutGetStoredBlock() NewRemoteBlockStore()) %v != %v", err, nil)
	}
	storeAPI := longtaillib.CreateBlockStoreAPI(remoteStore)

	b1, errno := storeBlockFromSeed(t, storeAPI, 0)
	if errno != 0 {
		t.Errorf("TestPutGetStoredBlock() storeBlock(t, storeAPI, 0) %d != %d", errno, 0)
	}
	b1Index := b1.GetBlockIndex()
	b1ChunkHashes := b1Index.GetChunkHashes()

	b2, errno := storeBlockFromSeed(t, storeAPI, 10)
	if errno != 0 {
		t.Errorf("TestPutGetStoredBlock() storeBlock(t, storeAPI, 10) %d != %d", errno, 0)
	}
	b2Index := b2.GetBlockIndex()
	b2ChunkHashes := b2Index.GetChunkHashes()

	b3, errno := storeBlockFromSeed(t, storeAPI, 20)
	if errno != 0 {
		t.Errorf("TestPutGetStoredBlock() storeBlock(t, storeAPI, 20) %d != %d", errno, 0)
	}
	b3Index := b3.GetBlockIndex()
	b3ChunkHashes := b3Index.GetChunkHashes()

	chunkHashes1 := []uint64{b1ChunkHashes[0], b1ChunkHashes[1], b3ChunkHashes[1], b3ChunkHashes[2]}
	chunkHashes2 := []uint64{b2ChunkHashes[0], b2ChunkHashes[1], b3ChunkHashes[0], b3ChunkHashes[1]}

	b1.Dispose()
	b2.Dispose()
	b3.Dispose()

	storeAPI.Dispose()

	remoteStore, err = NewRemoteBlockStore(
		jobs,
		blobStore,
		ReadWrite)
	if err != nil {
		t.Errorf("TestPutGetStoredBlock() NewRemoteBlockStore()) %v != %v", err, nil)
	}
	storeAPI = longtaillib.CreateBlockStoreAPI(remoteStore)

	existingContent, errno := getExistingContent(t, storeAPI, chunkHashes1, 0)
	if !existingContent.IsValid() {
		t.Errorf("TestRestoreStore() g.err %t != %t", existingContent.IsValid(), true)
	}

	if existingContent.GetBlockCount() != 2 {
		t.Errorf("TestRestoreStore() existingContent.GetBlockCount() %d != %d", existingContent.GetBlockCount(), 2)
	}

	if existingContent.GetChunkCount() != 6 {
		t.Errorf("TestRestoreStore() existingContent.GetChunkCount() %d != %d", existingContent.GetChunkCount(), 6)
	}

	existingContent, errno = getExistingContent(t, storeAPI, chunkHashes2, 0)
	if !existingContent.IsValid() {
		t.Errorf("TestRestoreStore() g.err %t != %t", existingContent.IsValid(), true)
	}

	if existingContent.GetBlockCount() != 2 {
		t.Errorf("TestRestoreStore() existingContent.GetBlockCount() %d != %d", existingContent.GetBlockCount(), 2)
	}

	if existingContent.GetChunkCount() != 6 {
		t.Errorf("TestRestoreStore() existingContent.GetChunkCount() %d != %d", existingContent.GetChunkCount(), 6)
	}

	b5, errno := storeBlockFromSeed(t, storeAPI, 30)
	if errno != 0 {
		t.Errorf("TestRestoreStore() storeBlock(t, storeAPI, 30) %d != %d", errno, 0)
	}
	b5Index := b5.GetBlockIndex()
	b5ChunkHashes := b5Index.GetChunkHashes()
	b5.Dispose()

	chunkHashes3 := append(chunkHashes2, b5ChunkHashes[1])

	existingContent.Dispose()
	storeAPI.Dispose()

	remoteStore, err = NewRemoteBlockStore(
		jobs,
		blobStore,
		ReadWrite)
	if err != nil {
		t.Errorf("TestRestoreStore() NewRemoteBlockStore()) %v != %v", err, nil)
	}
	storeAPI = longtaillib.CreateBlockStoreAPI(remoteStore)

	existingContent, errno = getExistingContent(t, storeAPI, chunkHashes3, 0)
	if !existingContent.IsValid() {
		t.Errorf("TestRestoreStore() g.err %t != %t", existingContent.IsValid(), true)
	}

	if existingContent.GetBlockCount() != 3 {
		t.Errorf("TestRestoreStore() existingContent.GetBlockCount() %d != %d", existingContent.GetBlockCount(), 3)
	}

	if existingContent.GetChunkCount() != 9 {
		t.Errorf("TestRestoreStore() existingContent.GetChunkCount() %d != %d", existingContent.GetChunkCount(), 9)
	}
	existingContent.Dispose()
	storeAPI.Dispose()
}

func createStoredBlock(chunkCount uint32, hashIdentifier uint32) (longtaillib.Longtail_StoredBlock, int) {
	blockHash := uint64(0xdeadbeef500177aa) + uint64(chunkCount)
	chunkHashes := make([]uint64, chunkCount)
	chunkSizes := make([]uint32, chunkCount)
	blockOffset := uint32(0)
	for index, _ := range chunkHashes {
		chunkHashes[index] = uint64(index+1) * 4711
		chunkSizes[index] = uint32(index+1) * 10
		blockOffset += uint32(chunkSizes[index])
	}
	blockData := make([]uint8, blockOffset)
	blockOffset = 0
	for chunkIndex, _ := range chunkHashes {
		for index := uint32(0); index < uint32(chunkSizes[chunkIndex]); index++ {
			blockData[blockOffset+index] = uint8(chunkIndex + 1)
		}
		blockOffset += uint32(chunkSizes[chunkIndex])
	}

	return longtaillib.CreateStoredBlock(
		blockHash,
		hashIdentifier,
		chunkCount+uint32(10000),
		chunkHashes,
		chunkSizes,
		blockData,
		false)
}

func storeBlock(blobClient BlobClient, storedBlock longtaillib.Longtail_StoredBlock, blockHashOffset uint64, parentPath string) uint64 {
	bytes, _ := longtaillib.WriteStoredBlockToBuffer(storedBlock)
	blockIndex := storedBlock.GetBlockIndex()
	storedBlockHash := blockIndex.GetBlockHash() + blockHashOffset
	path := GetBlockPath("chunks", storedBlockHash)
	if len(parentPath) > 0 {
		path = parentPath + "/" + path
	}
	blobObject, _ := blobClient.NewObject(path)
	blobObject.Write(bytes)
	return storedBlockHash
}

func TestBlockScanning(t *testing.T) {
	// Create stored blocks
	// Create/move stored blocks to faulty path
	// Scan and make sure we only get the blocks in the currect path
	blobStore, _ := NewTestBlobStore("")
	blobClient, _ := blobStore.NewClient(context.Background())

	goodBlockInCorrectPath, _ := generateStoredBlock(t, 7)
	goodBlockInCorrectPathHash := storeBlock(blobClient, goodBlockInCorrectPath, 0, "")

	badBlockInCorrectPath, _ := generateStoredBlock(t, 14)
	badBlockInCorrectPathHash := storeBlock(blobClient, badBlockInCorrectPath, 1, "")

	goodBlockInBadPath, _ := generateStoredBlock(t, 21)
	goodBlockInBadPathHash := storeBlock(blobClient, goodBlockInBadPath, 0, "chunks")

	badBlockInBatPath, _ := generateStoredBlock(t, 33)
	badBlockInBatPathHash := storeBlock(blobClient, badBlockInBatPath, 2, "chunks")

	jobs := longtaillib.CreateBikeshedJobAPI(uint32(runtime.NumCPU()), 0)
	defer jobs.Dispose()
	remoteStore, err := NewRemoteBlockStore(
		jobs,
		blobStore,
		Init)
	if err != nil {
		t.Errorf("TestPutGetStoredBlock() NewRemoteBlockStore()) %v != %v", err, nil)
	}
	storeAPI := longtaillib.CreateBlockStoreAPI(remoteStore)
	defer storeAPI.Dispose()

	b, errno := fetchBlockFromStore(t, storeAPI, goodBlockInCorrectPathHash)
	if errno != 0 {
		t.Errorf("TestBlockScanning() fetchBlockFromStore(t, storeAPI, goodBlockInCorrectPathHash) %d != %d", errno, 0)
	}
	b.Dispose()

	_, errno = fetchBlockFromStore(t, storeAPI, badBlockInCorrectPathHash)
	if errno != longtaillib.EBADF {
		t.Errorf("TestBlockScanning() fetchBlockFromStore(t, storeAPI, badBlockInCorrectPathHash) %d != %d", errno, longtaillib.ENOENT)
	}

	_, errno = fetchBlockFromStore(t, storeAPI, goodBlockInBadPathHash)
	if errno != longtaillib.ENOENT {
		t.Errorf("TestBlockScanning() fetchBlockFromStore(t, storeAPI, goodBlockInBadPathHash) %d != %d", errno, longtaillib.ENOENT)
	}

	_, errno = fetchBlockFromStore(t, storeAPI, badBlockInBatPathHash)
	if errno != longtaillib.ENOENT {
		t.Errorf("TestBlockScanning() fetchBlockFromStore(t, storeAPI, badBlockInBatPathHash) %d != %d", errno, longtaillib.ENOENT)
	}

	goodBlockInCorrectPathIndex := goodBlockInCorrectPath.GetBlockIndex()
	chunks := goodBlockInCorrectPathIndex.GetChunkHashes()
	badBlockInCorrectPathIndex := badBlockInCorrectPath.GetBlockIndex()
	chunks = append(chunks, badBlockInCorrectPathIndex.GetChunkHashes()...)
	goodBlockInBadPathIndex := goodBlockInBadPath.GetBlockIndex()
	chunks = append(chunks, goodBlockInBadPathIndex.GetChunkHashes()...)
	badBlockInBatPathIndex := badBlockInBatPath.GetBlockIndex()
	chunks = append(chunks, badBlockInBatPathIndex.GetChunkHashes()...)

	existingContent, errno := getExistingContent(t, storeAPI, chunks, 0)
	if errno != 0 {
		t.Errorf("TestBlockScanning() getExistingContent(t, storeAPI, chunkHashes, 0) %d != %d", errno, 0)
	}
	defer existingContent.Dispose()
	if len(existingContent.GetChunkHashes()) != len(goodBlockInCorrectPathIndex.GetChunkHashes()) {
		t.Errorf("TestBlockScanning() getExistingContent(t, storeAPI, chunks, 0) %d!= %d", len(existingContent.GetChunkHashes()), len(goodBlockInCorrectPathIndex.GetChunkHashes()))
	}
}
