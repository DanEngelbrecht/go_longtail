package longtaillib

import (
	"crypto/rand"
	"fmt"
	"runtime"
	"sync"
	"testing"
)

type testProgress struct {
	inited bool
	task   string
	t      *testing.T
}

func (p *testProgress) OnProgress(total uint32, current uint32) {
	if current == total {
		if p.inited {
			p.t.Logf("100%%")
			p.t.Logf(" Done\n")
		}
		return
	}
	if !p.inited {
		p.t.Logf("%s: ", p.task)
		p.inited = true
	}
	percentDone := (100 * current) / total
	p.t.Logf("%d%% ", percentDone)
}

func CreateProgress(t *testing.T, task string) Longtail_ProgressAPI {
	baseProgress := CreateProgressAPI(&testProgress{task: task, t: t})
	return CreateRateLimitedProgressAPI(baseProgress, 5)
}

type testPutBlockCompletionAPI struct {
	wg    sync.WaitGroup
	errno int
}

func (a *testPutBlockCompletionAPI) OnComplete(errno int) {
	a.errno = errno
	a.wg.Done()
}

type testGetBlockCompletionAPI struct {
	wg          sync.WaitGroup
	storedBlock Longtail_StoredBlock
	errno       int
}

func (a *testGetBlockCompletionAPI) OnComplete(storedBlock Longtail_StoredBlock, errno int) {
	a.storedBlock = storedBlock
	a.errno = errno
	a.wg.Done()
}

type testGetExistingContentCompletionAPI struct {
	wg         sync.WaitGroup
	storeIndex Longtail_StoreIndex
	errno      int
}

func (a *testGetExistingContentCompletionAPI) OnComplete(storeIndex Longtail_StoreIndex, errno int) {
	a.errno = errno
	a.storeIndex = storeIndex
	a.wg.Done()
}

type testLogger struct {
	t *testing.T
}

func (l *testLogger) OnLog(file string, function string, line int, level int, logFields []LogField, message string) {
	l.t.Logf("%d: %s", level, message)
}

type testAssert struct {
	t *testing.T
}

func (a *testAssert) OnAssert(expression string, file string, line int) {
	fmt.Printf("ASSERT: %s %s:%d", expression, file, line)
}

func TestDebugging(t *testing.T) {
	SetLogger(&testLogger{t: t})
	defer SetLogger(nil)
	SetAssert(&testAssert{t: t})
	defer SetAssert(nil)
	SetLogLevel(3)
}

func TestInMemStorage(t *testing.T) {
	SetLogger(&testLogger{t: t})
	defer SetLogger(nil)
	SetAssert(&testAssert{t: t})
	defer SetAssert(nil)
	SetLogLevel(1)

	storageAPI := CreateInMemStorageAPI()
	defer storageAPI.Dispose()
	myString := "my string"
	errno := storageAPI.WriteToStorage("folder", "file", []byte(myString))
	expected := 0
	if errno != 0 {
		t.Errorf("WriteToStorage() %d != %d", errno, expected)
	}

	rbytes, errno := storageAPI.ReadFromStorage("folder", "file")
	testString := string(rbytes)
	if myString != myString {
		t.Errorf("ReadFromStorage() %s != %s", rbytes, testString)
	}
}

func TestAPICreate(t *testing.T) {
	SetLogger(&testLogger{t: t})
	defer SetLogger(nil)
	SetAssert(&testAssert{t: t})
	defer SetAssert(nil)
	SetLogLevel(1)

	blake2 := CreateBlake2HashAPI()
	defer blake2.Dispose()

	blake3 := CreateBlake3HashAPI()
	defer blake3.Dispose()

	meow := CreateMeowHashAPI()
	defer meow.Dispose()

	lz4 := CreateLZ4CompressionAPI()
	defer lz4.Dispose()

	brotli := CreateBrotliCompressionAPI()
	defer brotli.Dispose()

	zstd := CreateZStdCompressionAPI()
	defer zstd.Dispose()

	jobAPI := CreateBikeshedJobAPI(uint32(runtime.NumCPU()), 0)
	defer jobAPI.Dispose()

	compressionRegistry := CreateZStdCompressionRegistry()
	defer compressionRegistry.Dispose()
}

func createStoredBlock(chunkCount uint32, hashIdentifier uint32) (Longtail_StoredBlock, int) {
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

	return CreateStoredBlock(
		blockHash,
		hashIdentifier,
		chunkCount+uint32(10000),
		chunkHashes,
		chunkSizes,
		blockData,
		false)
}

func validateStoredBlock(t *testing.T, storedBlock Longtail_StoredBlock, hashIdentifier uint32) {
	if storedBlock.cStoredBlock == nil {
		t.Errorf("validateStoredBlock() %p == %p", storedBlock, Longtail_StoredBlock{cStoredBlock: nil})
	}
	blockIndex := storedBlock.GetBlockIndex()
	if blockIndex.GetHashIdentifier() != hashIdentifier {
		t.Errorf("validateStoredBlock() %d == %d", blockIndex.GetHashIdentifier(), hashIdentifier)
	}
	chunkCount := blockIndex.GetChunkCount()
	if blockIndex.GetBlockHash() != uint64(0xdeadbeef500177aa)+uint64(chunkCount) {
		t.Errorf("validateStoredBlock() %q != %q", uint64(0xdeadbeef500177aa)+uint64(chunkCount), blockIndex.GetBlockHash())
	}
	if blockIndex.GetTag() != chunkCount+uint32(10000) {
		t.Errorf("validateStoredBlock() %q != %q", chunkCount+uint32(10000), blockIndex.GetTag())

	}
	chunkHashes := blockIndex.GetChunkHashes()
	if uint32(len(chunkHashes)) != chunkCount {
		t.Errorf("validateStoredBlock() %q != %q", chunkCount, uint32(len(chunkHashes)))
	}
	chunkSizes := blockIndex.GetChunkSizes()
	if uint32(len(chunkSizes)) != chunkCount {
		t.Errorf("validateStoredBlock() %q != %q", chunkCount, uint32(len(chunkSizes)))
	}
	blockOffset := uint32(0)
	for index, _ := range chunkHashes {
		if chunkHashes[index] != uint64(index+1)*4711 {
			t.Errorf("validateStoredBlock() %q != %q", uint64(index)*4711, chunkHashes[index])
		}
		if chunkSizes[index] != uint32(index+1)*10 {
			t.Errorf("validateStoredBlock() %q != %q", uint32(index)*10, chunkSizes[index])
		}
		blockOffset += uint32(chunkSizes[index])
	}
	blockData := storedBlock.GetChunksBlockData()
	if uint32(len(blockData)) != blockOffset {
		t.Errorf("validateStoredBlock() %q != %q", uint32(len(blockData)), blockOffset)
	}
	blockOffset = 0
	for chunkIndex, _ := range chunkHashes {
		for index := uint32(0); index < uint32(chunkSizes[chunkIndex]); index++ {
			if blockData[blockOffset+index] != uint8(chunkIndex+1) {
				t.Errorf("validateStoredBlock() %q != %q", uint8(chunkIndex+1), blockData[blockOffset+index])
			}
		}
		blockOffset += uint32(chunkSizes[chunkIndex])
	}
}

func TestStoredblock(t *testing.T) {
	SetLogger(&testLogger{t: t})
	defer SetLogger(nil)
	SetAssert(&testAssert{t: t})
	defer SetAssert(nil)
	SetLogLevel(1)

	storedBlock, errno := createStoredBlock(2, 0xdeadbeef)
	if errno != 0 {
		t.Errorf("CreateStoredBlock() %d != %d", errno, 0)
	}
	validateStoredBlock(t, storedBlock, 0xdeadbeef)
}

func Test_ReadWriteStoredBlockBuffer(t *testing.T) {
	SetLogger(&testLogger{t: t})
	defer SetLogger(nil)
	SetAssert(&testAssert{t: t})
	defer SetAssert(nil)
	SetLogLevel(1)

	originalBlock, errno := createStoredBlock(2, 0xdeadbeef)
	if errno != 0 {
		t.Errorf("createStoredBlock() %d != %d", errno, 0)
	}

	storedBlockData, errno := WriteStoredBlockToBuffer(originalBlock)
	if errno != 0 {
		t.Errorf("WriteStoredBlockToBuffer() %d != %d", errno, 0)
	}
	originalBlock.Dispose()

	copyBlock, errno := ReadStoredBlockFromBuffer(storedBlockData)

	if errno != 0 {
		t.Errorf("InitStoredBlockFromData() %d != %d", errno, 0)
	}
	defer copyBlock.Dispose()
	validateStoredBlock(t, copyBlock, 0xdeadbeef)
}

func TestFSBlockStore(t *testing.T) {
	SetLogger(&testLogger{t: t})
	defer SetLogger(nil)
	SetAssert(&testAssert{t: t})
	defer SetAssert(nil)
	SetLogLevel(1)

	storageAPI := CreateInMemStorageAPI()
	defer storageAPI.Dispose()
	jobAPI := CreateBikeshedJobAPI(uint32(runtime.NumCPU()), 0)
	defer jobAPI.Dispose()
	blockStoreAPI := CreateFSBlockStore(jobAPI, storageAPI, "content", 8388608, 1024)
	defer blockStoreAPI.Dispose()
	blake3 := CreateBlake3HashAPI()
	defer blake3.Dispose()

	block1, errno := createStoredBlock(1, blake3.GetIdentifier())
	if errno != 0 {
		t.Errorf("TestFSBlockStore() createStoredBlock() %d != %d", errno, 0)
	}
	defer block1.Dispose()

	block2, errno := createStoredBlock(5, blake3.GetIdentifier())
	if errno != 0 {
		t.Errorf("TestFSBlockStore() createStoredBlock() %d != %d", errno, 0)
	}
	defer block2.Dispose()

	block3, errno := createStoredBlock(9, blake3.GetIdentifier())
	if errno != 0 {
		t.Errorf("TestFSBlockStore() createStoredBlock() %d != %d", errno, 0)
	}
	defer block3.Dispose()

	storedBlock1Index := block1.GetBlockIndex()
	storedBlock1Hash := storedBlock1Index.GetBlockHash()
	getStoredBlockComplete := &testGetBlockCompletionAPI{}
	getStoredBlockComplete.wg.Add(1)
	errno = blockStoreAPI.GetStoredBlock(storedBlock1Hash, CreateAsyncGetStoredBlockAPI(getStoredBlockComplete))
	if errno != ENOENT {
		t.Errorf("TestFSBlockStore() GetStoredBlock() %d == %d", errno, ENOENT)
		getStoredBlockComplete.wg.Done()
	}
	getStoredBlockComplete.wg.Done()
	nullBlock := Longtail_StoredBlock{}
	if getStoredBlockComplete.storedBlock != nullBlock {
		t.Errorf("TestFSBlockStore() getStoredBlockComplete.storedBlock %p != %p", getStoredBlockComplete.storedBlock, nullBlock)
	}

	putStoredBlockComplete := &testPutBlockCompletionAPI{}
	putStoredBlockComplete.wg.Add(1)
	errno = blockStoreAPI.PutStoredBlock(block1, CreateAsyncPutStoredBlockAPI(putStoredBlockComplete))
	if errno != 0 {
		t.Errorf("TestFSBlockStore() PutStoredBlock() %d != %d", errno, 0)
		putStoredBlockComplete.wg.Done()
	}
	putStoredBlockComplete.wg.Wait()
	if putStoredBlockComplete.errno != 0 {
		t.Errorf("TestFSBlockStore() putStoredBlockComplete.errno %d != %d", putStoredBlockComplete.errno, 0)
	}

	getStoredBlockComplete.wg.Add(1)
	errno = blockStoreAPI.GetStoredBlock(storedBlock1Hash, CreateAsyncGetStoredBlockAPI(getStoredBlockComplete))
	if errno != 0 {
		t.Errorf("TestFSBlockStore() GetStoredBlock() %d != %d", errno, 0)
		getStoredBlockComplete.wg.Done()
	}
	getStoredBlockComplete.wg.Wait()
	if getStoredBlockComplete.errno != 0 {
		t.Errorf("TestFSBlockStore() getStoredBlockComplete.errno %d != %d", getStoredBlockComplete.errno, 0)
	}
	storedBlock1 := getStoredBlockComplete.storedBlock
	getStoredBlockComplete.storedBlock = nullBlock
	if storedBlock1 == nullBlock {
		t.Errorf("TestFSBlockStore() HasStoredBlock() %v != %v", storedBlock1, nullBlock)
	}
	defer storedBlock1.Dispose()
	validateStoredBlock(t, storedBlock1, blake3.GetIdentifier())

	putStoredBlockComplete.wg.Add(1)
	errno = blockStoreAPI.PutStoredBlock(block2, CreateAsyncPutStoredBlockAPI(putStoredBlockComplete))
	if errno != 0 {
		t.Errorf("TestFSBlockStore() PutStoredBlock() %d != %d", errno, 0)
		putStoredBlockComplete.wg.Done()
	}
	putStoredBlockComplete.wg.Wait()
	if putStoredBlockComplete.errno != 0 {
		t.Errorf("TestFSBlockStore() putStoredBlockComplete.errno %d != %d", putStoredBlockComplete.errno, 0)
	}

	storedBlock2Index := block2.GetBlockIndex()
	storedBlock2Hash := storedBlock2Index.GetBlockHash()
	getStoredBlockComplete.wg.Add(1)
	errno = blockStoreAPI.GetStoredBlock(storedBlock2Hash, CreateAsyncGetStoredBlockAPI(getStoredBlockComplete))
	if errno != 0 {
		t.Errorf("TestFSBlockStore() HasStoredBlock() %d != %d", errno, 0)
		getStoredBlockComplete.wg.Done()
	}
	getStoredBlockComplete.wg.Wait()
	if getStoredBlockComplete.errno != 0 {
		t.Errorf("TestFSBlockStore() getStoredBlockComplete.errno %d != %d", getStoredBlockComplete.errno, 0)
	}
	storedBlock2 := getStoredBlockComplete.storedBlock
	getStoredBlockComplete.storedBlock = nullBlock
	if storedBlock2 == nullBlock {
		t.Errorf("TestFSBlockStore() HasStoredBlock() %v != %v", storedBlock2, nullBlock)
	}
	defer storedBlock2.Dispose()
	validateStoredBlock(t, storedBlock2, blake3.GetIdentifier())

	putStoredBlockComplete.wg.Add(1)
	errno = blockStoreAPI.PutStoredBlock(block3, CreateAsyncPutStoredBlockAPI(putStoredBlockComplete))
	if errno != 0 {
		t.Errorf("TestFSBlockStore() PutStoredBlock() %d != %d", errno, 0)
		putStoredBlockComplete.wg.Done()
	}
	putStoredBlockComplete.wg.Wait()
	if putStoredBlockComplete.errno != 0 {
		t.Errorf("TestFSBlockStore() putStoredBlockComplete.errno %d != %d", putStoredBlockComplete.errno, 0)
	}

	storedBlock3Index := block3.GetBlockIndex()
	storedBlock3Hash := storedBlock3Index.GetBlockHash()
	getStoredBlockComplete.wg.Add(1)
	errno = blockStoreAPI.GetStoredBlock(storedBlock3Hash, CreateAsyncGetStoredBlockAPI(getStoredBlockComplete))
	if errno != 0 {
		t.Errorf("TestFSBlockStore() GetStoredBlock() %d != %d", errno, 0)
		getStoredBlockComplete.wg.Done()
	}
	getStoredBlockComplete.wg.Wait()
	if getStoredBlockComplete.errno != 0 {
		t.Errorf("TestFSBlockStore() getStoredBlockComplete.errno %d != %d", getStoredBlockComplete.errno, 0)
	}
	storedBlock3 := getStoredBlockComplete.storedBlock
	getStoredBlockComplete.storedBlock = nullBlock
	if storedBlock3 == nullBlock {
		t.Errorf("TestFSBlockStore() HasStoredBlock() %v != %v", storedBlock3, nullBlock)
	}
	defer storedBlock3.Dispose()
	validateStoredBlock(t, storedBlock3, blake3.GetIdentifier())

	getStoredBlockComplete.wg.Add(1)
	errno = blockStoreAPI.GetStoredBlock(storedBlock2Hash, CreateAsyncGetStoredBlockAPI(getStoredBlockComplete))
	if errno != 0 {
		t.Errorf("TestFSBlockStore() GetStoredBlock() %d != %d", errno, 0)
		getStoredBlockComplete.wg.Done()
	}
	getStoredBlockComplete.wg.Wait()
	if getStoredBlockComplete.errno != 0 {
		t.Errorf("TestFSBlockStore() getStoredBlockComplete.errno %d != %d", getStoredBlockComplete.errno, 0)
	}
	storedBlock2Again := getStoredBlockComplete.storedBlock
	getStoredBlockComplete.storedBlock = nullBlock
	if storedBlock2Again == nullBlock {
		t.Errorf("TestFSBlockStore() HasStoredBlock() %v != %v", storedBlock2Again, nullBlock)
	}
	defer storedBlock2Again.Dispose()
	validateStoredBlock(t, storedBlock2Again, blake3.GetIdentifier())
}

type TestBlockStore struct {
	blocks        map[uint64]Longtail_StoredBlock
	blockStoreAPI Longtail_BlockStoreAPI
	lock          sync.Mutex
	stats         [Longtail_BlockStoreAPI_StatU64_Count]uint64
	didClose      bool
}

func (b *TestBlockStore) PutStoredBlock(
	storedBlock Longtail_StoredBlock,
	asyncCompleteAPI Longtail_AsyncPutStoredBlockAPI) int {
	b.lock.Lock()
	b.stats[Longtail_BlockStoreAPI_StatU64_PutStoredBlock_Count] += 1
	defer b.lock.Unlock()
	blockIndex := storedBlock.GetBlockIndex()
	blockHash := blockIndex.GetBlockHash()
	if _, ok := b.blocks[blockHash]; ok {
		return 0
	}
	blockCopy, errno := CreateStoredBlock(
		blockHash,
		blockIndex.GetHashIdentifier(),
		blockIndex.GetTag(),
		blockIndex.GetChunkHashes(),
		blockIndex.GetChunkSizes(),
		storedBlock.GetChunksBlockData(),
		false)
	if errno == 0 {
		b.blocks[blockHash] = blockCopy
		asyncCompleteAPI.OnComplete(0)
		return 0
	}
	asyncCompleteAPI.OnComplete(errno)
	return 0
}

func (b *TestBlockStore) PreflightGet(chunkHashes []uint64) int {
	b.stats[Longtail_BlockStoreAPI_StatU64_PreflightGet_Count] += 1
	return 0
}

func (b *TestBlockStore) GetStoredBlock(
	blockHash uint64,
	asyncCompleteAPI Longtail_AsyncGetStoredBlockAPI) int {
	b.lock.Lock()
	b.stats[Longtail_BlockStoreAPI_StatU64_GetStoredBlock_Count] += 1
	defer b.lock.Unlock()
	if storedBlock, ok := b.blocks[blockHash]; ok {
		blockIndex := storedBlock.GetBlockIndex()
		blockCopy, errno := CreateStoredBlock(
			blockIndex.GetBlockHash(),
			blockIndex.GetHashIdentifier(),
			blockIndex.GetTag(),
			blockIndex.GetChunkHashes(),
			blockIndex.GetChunkSizes(),
			storedBlock.GetChunksBlockData(),
			false)
		if errno == 0 {
			asyncCompleteAPI.OnComplete(blockCopy, errno)
			return 0
		}
	}
	asyncCompleteAPI.OnComplete(Longtail_StoredBlock{}, ENOENT)
	return 0
}

func (b *TestBlockStore) GetIndexSync() (Longtail_StoreIndex, int) {
	b.lock.Lock()
	defer b.lock.Unlock()
	blockCount := len(b.blocks)
	blockIndexes := make([]Longtail_BlockIndex, blockCount)
	arrayIndex := 0
	for _, value := range b.blocks {
		blockIndexes[arrayIndex] = value.GetBlockIndex()
		arrayIndex++
	}
	sIndex, errno := CreateStoreIndexFromBlocks(blockIndexes)
	return sIndex, errno
}

func (b *TestBlockStore) GetExistingContent(
	chunkHashes []uint64,
	minBlockUsagePercent uint32,
	asyncCompleteAPI Longtail_AsyncGetExistingContentAPI) int {
	b.stats[Longtail_BlockStoreAPI_StatU64_GetExistingContent_Count] += 1
	sIndex, errno := b.GetIndexSync()
	if errno != 0 {
		asyncCompleteAPI.OnComplete(Longtail_StoreIndex{}, errno)
		return 0
	}
	defer sIndex.Dispose()

	sExistingIndex, errno := GetExistingStoreIndex(
		sIndex,
		chunkHashes,
		minBlockUsagePercent)
	if errno != 0 {
		asyncCompleteAPI.OnComplete(Longtail_StoreIndex{}, errno)
		return 0
	}
	asyncCompleteAPI.OnComplete(sExistingIndex, 0)
	return 0
}

// GetStats ...
func (b *TestBlockStore) GetStats() (BlockStoreStats, int) {
	b.stats[Longtail_BlockStoreAPI_StatU64_GetStats_Count] += 1
	var stats BlockStoreStats
	for i := 0; i < Longtail_BlockStoreAPI_StatU64_Count; i++ {
		stats.StatU64[i] = b.stats[i]
	}
	return stats, 0
}

func (b *TestBlockStore) Flush(asyncCompleteAPI Longtail_AsyncFlushAPI) int {
	b.stats[Longtail_BlockStoreAPI_StatU64_Flush_Count] += 1
	asyncCompleteAPI.OnComplete(0)
	return 0
}

func (b *TestBlockStore) Close() {
	b.didClose = true
}

func TestBlockStoreProxy(t *testing.T) {
	SetLogger(&testLogger{t: t})
	defer SetLogger(nil)
	SetAssert(&testAssert{t: t})
	defer SetAssert(nil)
	SetLogLevel(1)

	blockStore := &TestBlockStore{blocks: make(map[uint64]Longtail_StoredBlock), didClose: false}
	blockStoreProxy := CreateBlockStoreAPI(blockStore)
	blockStore.blockStoreAPI = blockStoreProxy
	defer blockStoreProxy.Dispose()

	storedBlock, errno := createStoredBlock(2, 0xdeadbeef)
	if errno != 0 {
		t.Errorf("TestBlockStoreProxy() createStoredBlock() %d != %d", errno, 0)
	}
	defer storedBlock.Dispose()

	putStoredBlockComplete := &testPutBlockCompletionAPI{}
	putStoredBlockComplete.wg.Add(1)
	errno = blockStoreProxy.PutStoredBlock(storedBlock, CreateAsyncPutStoredBlockAPI(putStoredBlockComplete))
	if errno != 0 {
		t.Errorf("TestBlockStoreProxy() PutStoredBlock() %d != %d", errno, 0)
		putStoredBlockComplete.wg.Done()
	}
	putStoredBlockComplete.wg.Wait()
	if putStoredBlockComplete.errno != 0 {
		t.Errorf("TestBlockStoreProxy() putStoredBlockComplete.errno %d != %d", putStoredBlockComplete.errno, 0)
	}

	getStoredBlockComplete := &testGetBlockCompletionAPI{}
	storedBlockIndex := storedBlock.GetBlockIndex()
	getStoredBlockComplete.wg.Add(1)
	errno = blockStoreProxy.GetStoredBlock(storedBlockIndex.GetBlockHash(), CreateAsyncGetStoredBlockAPI(getStoredBlockComplete))
	if errno != 0 {
		t.Errorf("TestBlockStoreProxy() GetStoredBlock() %d!= %d", errno, 0)
		getStoredBlockComplete.wg.Done()
	}
	getStoredBlockComplete.wg.Wait()
	if getStoredBlockComplete.errno != 0 {
		t.Errorf("TestFSBlockStore() getStoredBlockComplete.errno %d != %d", getStoredBlockComplete.errno, 0)
	}
	getBlock := getStoredBlockComplete.storedBlock
	getStoredBlockComplete.storedBlock = Longtail_StoredBlock{}
	defer getBlock.Dispose()
	validateStoredBlock(t, getBlock, 0xdeadbeef)

	stats, errno := blockStoreProxy.GetStats()
	if errno != 0 {
		t.Errorf("TestBlockStoreProxy() GetStats() %d != %d", errno, 0)
	}
	if stats.StatU64[Longtail_BlockStoreAPI_StatU64_GetStoredBlock_Count] != 1 {
		t.Errorf("TestBlockStoreProxy() stats.BlocksGetCount %d != %d", stats.StatU64[Longtail_BlockStoreAPI_StatU64_GetStoredBlock_Count], 1)
	}
	if stats.StatU64[Longtail_BlockStoreAPI_StatU64_PutStoredBlock_Count] != 1 {
		t.Errorf("TestBlockStoreProxy() stats.BlocksPutCount %d != %d", stats.StatU64[Longtail_BlockStoreAPI_StatU64_PutStoredBlock_Count], 1)
	}

	blockStoreProxy.Dispose()
}

type testPathFilter struct {
}

func (p *testPathFilter) Include(rootPath string, assetFolder string, assetName string, isDir bool, size uint64, permissions uint16) bool {
	return true
}

func TestBlockStoreProxyFull(t *testing.T) {
	storageAPI := createFilledStorage("content")
	defer storageAPI.Dispose()
	hashAPI := CreateBlake3HashAPI()
	defer hashAPI.Dispose()
	chunkerAPI := CreateHPCDCChunkerAPI()
	defer chunkerAPI.Dispose()
	jobAPI := CreateBikeshedJobAPI(uint32(runtime.NumCPU()), 0)
	defer jobAPI.Dispose()
	testBlockStore := &TestBlockStore{blocks: make(map[uint64]Longtail_StoredBlock)}
	blockStoreAPI := CreateBlockStoreAPI(testBlockStore)
	defer blockStoreAPI.Dispose()

	pathFilter := CreatePathFilterAPI(&testPathFilter{})

	fileInfos, errno := GetFilesRecursively(storageAPI, pathFilter, "content")
	if errno != 0 {
		t.Errorf("TestBlockStoreProxyFull() GetFilesRecursively() %q != %v", errno, 0)
	}
	defer fileInfos.Dispose()
	tags := make([]uint32, fileInfos.GetFileCount())
	versionIndex, errno := CreateVersionIndex(
		storageAPI,
		hashAPI,
		chunkerAPI,
		jobAPI,
		nil,
		"content",
		fileInfos,
		tags,
		32768)
	if errno != 0 {
		t.Errorf("TestBlockStoreProxyFull() CreateVersionIndex() %q != %v", errno, 0)
	}
	defer versionIndex.Dispose()

	getExistingContentComplete := &testGetExistingContentCompletionAPI{}
	getExistingContentComplete.wg.Add(1)
	errno = blockStoreAPI.GetExistingContent(versionIndex.GetChunkHashes(), 0, CreateAsyncGetExistingContentAPI(getExistingContentComplete))
	if errno != 0 {
		t.Errorf("TestBlockStoreProxyFull() blockStoreAPI.GetExistingContent() %d != %d", errno, 0)
		getExistingContentComplete.wg.Done()
	}
	getExistingContentComplete.wg.Wait()
	blockStoreIndex := getExistingContentComplete.storeIndex
	defer blockStoreIndex.Dispose()

	missingStoreIndex, errno := CreateMissingContent(
		hashAPI,
		blockStoreIndex,
		versionIndex,
		32768*2,
		8)
	if errno != 0 {
		t.Errorf("TestBlockStoreProxyFull() CreateMissingContent() %d != %d", errno, 0)
		getExistingContentComplete.wg.Done()
	}
	defer missingStoreIndex.Dispose()

	errno = WriteContent(
		storageAPI,
		blockStoreAPI,
		jobAPI,
		nil,
		missingStoreIndex,
		versionIndex,
		"content")
	if errno != 0 {
		t.Errorf("TestBlockStoreProxyFull() WriteContent() %d != %d", errno, 0)
	}
}

func randomArray(size int) []byte {
	r := make([]byte, size)
	rand.Read(r)
	return r
}

func createFilledStorage(rootPath string) Longtail_StorageAPI {
	storageAPI := CreateInMemStorageAPI()
	storageAPI.WriteToStorage(rootPath, "first_folder/my_file.txt", []byte("the content of my_file"))
	storageAPI.WriteToStorage(rootPath, "second_folder/my_second_file.txt", []byte("second file has different content than my_file"))
	storageAPI.WriteToStorage(rootPath, "top_level.txt", []byte("the top level file is also a text file with dummy content"))
	storageAPI.WriteToStorage(rootPath, "first_folder/empty/file/deeply/nested/file/in/lots/of/nests.txt", []byte{})
	storageAPI.WriteToStorage(rootPath, "bin/small.bin", randomArray(8192))
	storageAPI.WriteToStorage(rootPath, "bin/huge.bin", randomArray(65535*16))
	storageAPI.WriteToStorage(rootPath, "bin/medium.bin", randomArray(32768))
	storageAPI.WriteToStorage(rootPath, "bin/large.bin", randomArray(65535))
	return storageAPI
}

func TestGetFileRecursively(t *testing.T) {
	storageAPI := createFilledStorage("content")
	fileInfos, errno := GetFilesRecursively(storageAPI, Longtail_PathFilterAPI{}, "content")
	if errno != 0 {
		t.Errorf("TestGetFileRecursively() GetFilesRecursively() %d != %d", errno, 0)
	}
	defer fileInfos.Dispose()
	fileCount := fileInfos.GetFileCount()
	if fileCount != 19 {
		t.Errorf("TestGetFileRecursively() GetFileCount() %d != %d", fileCount, 19)
	}
	fileSizes := fileInfos.GetFileSizes()
	if len(fileSizes) != int(fileCount) {
		t.Errorf("TestGetFileRecursively() GetFileSizes() %d != %d", len(fileSizes), fileCount)
	}
	permissions := fileInfos.GetFilePermissions()
	if len(permissions) != int(fileCount) {
		t.Errorf("TestGetFileRecursively() GetFilePermissions() %d != %d", len(permissions), fileCount)
	}
	path := fileInfos.GetPath(0)
	if path != "first_folder/" {
		t.Errorf("TestGetFileRecursively() GetPaths().GetPath() %s != %s", path, "first_folder/")
	}
}

func TestCreateVersionIndex(t *testing.T) {
	storageAPI := createFilledStorage("content")
	fileInfos, errno := GetFilesRecursively(storageAPI, Longtail_PathFilterAPI{}, "content")
	if errno != 0 {
		t.Errorf("TestCreateVersionIndex() GetFilesRecursively() %d != %d", errno, 0)
	}
	hashAPI := CreateBlake2HashAPI()
	defer hashAPI.Dispose()
	chunkerAPI := CreateHPCDCChunkerAPI()
	defer chunkerAPI.Dispose()
	jobAPI := CreateBikeshedJobAPI(uint32(runtime.NumCPU()), 0)
	defer jobAPI.Dispose()

	compressionTypes := make([]uint32, fileInfos.GetFileCount())

	versionIndex, errno := CreateVersionIndex(
		storageAPI,
		hashAPI,
		chunkerAPI,
		jobAPI,
		nil,
		"content",
		fileInfos,
		compressionTypes,
		32768)

	if errno != 0 {
		t.Errorf("TestCreateVersionIndex() CreateVersionIndex() %d != %d", errno, 0)
	}
	defer versionIndex.Dispose()
	if versionIndex.GetHashIdentifier() != hashAPI.GetIdentifier() {
		t.Errorf("TestCreateVersionIndex() GetHashIdentifier() %d != %d", versionIndex.GetHashIdentifier(), hashAPI.GetIdentifier())
	}
	if versionIndex.GetAssetCount() != fileInfos.GetFileCount() {
		t.Errorf("TestCreateVersionIndex() GetAssetCount() %d != %d", versionIndex.GetAssetCount(), fileInfos.GetFileCount())
	}
}

func TestRewriteVersion(t *testing.T) {
	storageAPI := createFilledStorage("content")
	fileInfos, errno := GetFilesRecursively(storageAPI, Longtail_PathFilterAPI{}, "content")
	if errno != 0 {
		t.Errorf("TestRewriteVersion() GetFilesRecursively() %d != %d", errno, 0)
	}
	hashAPI := CreateBlake2HashAPI()
	defer hashAPI.Dispose()
	chunkerAPI := CreateHPCDCChunkerAPI()
	defer chunkerAPI.Dispose()
	jobAPI := CreateBikeshedJobAPI(uint32(runtime.NumCPU()), 0)
	defer jobAPI.Dispose()

	compressionTypes := make([]uint32, fileInfos.GetFileCount())

	createVersionProgress := CreateProgress(t, "CreateVersionIndex")
	versionIndex, errno := CreateVersionIndex(
		storageAPI,
		hashAPI,
		chunkerAPI,
		jobAPI,
		&createVersionProgress,
		"content",
		fileInfos,
		compressionTypes,
		32768)
	if errno != 0 {
		t.Errorf("TestRewriteVersion() CreateVersionIndex() %d != %d", errno, 0)
	}

	storeIndex, errno := CreateStoreIndex(
		hashAPI,
		versionIndex,
		65536,
		4096)
	if errno != 0 {
		t.Errorf("TestRewriteVersion() CreateStoreIndex() %d != %d", errno, 0)
	}
	defer storeIndex.Dispose()
	blockStorageAPI := CreateFSBlockStore(jobAPI, storageAPI, "block_store", 65536, 4096)
	defer blockStorageAPI.Dispose()
	compressionRegistry := CreateZStdCompressionRegistry()
	compressionRegistry.Dispose()
	writeContentProgress := CreateProgress(t, "WriteContent")
	defer writeContentProgress.Dispose()

	errno = WriteContent(
		storageAPI,
		blockStorageAPI,
		jobAPI,
		&writeContentProgress,
		storeIndex,
		versionIndex,
		"content")
	if errno != 0 {
		t.Errorf("TestRewriteVersion() WriteContent() %d != %d", errno, 0)
	}

	getExistingContentComplete := &testGetExistingContentCompletionAPI{}
	getExistingContentComplete.wg.Add(1)
	errno = blockStorageAPI.GetExistingContent(versionIndex.GetChunkHashes(), 0, CreateAsyncGetExistingContentAPI(getExistingContentComplete))
	if errno != 0 {
		t.Errorf("TestBlockStoreProxyFull() blockStoreAPI.GetExistingContent() %d != %d", errno, 0)
		getExistingContentComplete.wg.Done()
	}
	getExistingContentComplete.wg.Wait()
	existingStoreIndex := getExistingContentComplete.storeIndex
	defer existingStoreIndex.Dispose()

	writeVersionProgress2 := CreateProgress(t, "WriteVersion")
	errno = WriteVersion(
		blockStorageAPI,
		storageAPI,
		jobAPI,
		&writeVersionProgress2,
		existingStoreIndex,
		versionIndex,
		"content_copy",
		true)
	if errno != 0 {
		t.Errorf("TestRewriteVersion() WriteVersion() %d != %d", errno, 0)
	}
}
