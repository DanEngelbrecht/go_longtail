package longtailstorelib

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/DanEngelbrecht/golongtail/longtaillib"
)

type testBlob struct {
	path string
	data []byte
}

type testBlobStore struct {
	blobs      map[string]*testBlob
	blobsMutex sync.RWMutex
	prefix     string
}

type testBlobClient struct {
	store *testBlobStore
}

type testBlobObject struct {
	client *testBlobClient
	path   string
}

// NewTestBlobStore ...
func NewTestBlobStore(prefix string) (BlobStore, error) {
	s := &testBlobStore{prefix: prefix, blobs: make(map[string]*testBlob)}
	return s, nil
}

func (blobStore *testBlobStore) NewClient(ctx context.Context) (BlobClient, error) {
	return &testBlobClient{store: blobStore}, nil
}

func (blobStore *testBlobStore) String() string {
	return "teststore"
}

func (blobClient *testBlobClient) NewObject(filepath string) (BlobObject, error) {
	return &testBlobObject{client: blobClient, path: filepath}, nil
}

func (blobClient *testBlobClient) GetObjects(pathPrefix string) ([]BlobProperties, error) {
	blobClient.store.blobsMutex.RLock()
	properties := make([]BlobProperties, 0)
	i := 0
	for key, blob := range blobClient.store.blobs {
		if strings.HasPrefix(key, pathPrefix) {
			properties = append(properties, BlobProperties{Name: key, Size: int64(len(blob.data))})
		}
		i++
	}
	blobClient.store.blobsMutex.RUnlock()
	//	time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)
	return properties, nil
}

func (blobClient *testBlobClient) Close() {
}

func (blobClient *testBlobClient) String() string {
	return "teststore"
}

func (blobObject *testBlobObject) Exists() (bool, error) {
	blobObject.client.store.blobsMutex.RLock()
	_, exists := blobObject.client.store.blobs[blobObject.path]
	blobObject.client.store.blobsMutex.RUnlock()
	//	time.Sleep(time.Duration(rand.Intn(10)+2) * time.Millisecond)
	return exists, nil
}

func (blobObject *testBlobObject) Read() ([]byte, error) {
	blobObject.client.store.blobsMutex.RLock()
	blob, exists := blobObject.client.store.blobs[blobObject.path]
	blobObject.client.store.blobsMutex.RUnlock()
	if !exists {
		return nil, nil
	}
	//	time.Sleep(time.Duration(rand.Intn(3)+len(blob.data)+2) * time.Millisecond)
	return blob.data, nil
}

func (blobObject *testBlobObject) Write(data []byte) (bool, error) {
	blobObject.client.store.blobsMutex.Lock()

	blob, exists := blobObject.client.store.blobs[blobObject.path]

	if !exists {
		blob = &testBlob{path: blobObject.path, data: data}
		blobObject.client.store.blobs[blobObject.path] = blob
		blobObject.client.store.blobsMutex.Unlock()
		//		time.Sleep(time.Duration(rand.Intn(3)+len(data)+3) * time.Millisecond)
		return true, nil
	}
	blob.data = data
	blobObject.client.store.blobsMutex.Unlock()
	//	time.Sleep((time.Duration(len(data)) + 3) * time.Millisecond)
	return true, nil
}

func (blobObject *testBlobObject) Delete() error {
	blobObject.client.store.blobsMutex.Lock()
	delete(blobObject.client.store.blobs, blobObject.path)
	blobObject.client.store.blobsMutex.Unlock()
	//	time.Sleep(time.Duration(rand.Intn(3)+rand.Intn(10)+2) * time.Millisecond)
	return nil
}

func generateStoredBlock(t *testing.T, seed uint8) (longtaillib.Longtail_StoredBlock, int) {
	chunkHashes := []uint64{uint64(seed) + 1, uint64(seed) + 2, uint64(seed) + 3}
	chunkSizes := []uint32{uint32(seed) + 10, uint32(seed) + 20, uint32(seed) + 30}

	blockDataLen := (int)(chunkSizes[0] + chunkSizes[1] + chunkSizes[2])
	blockData := make([]uint8, blockDataLen)
	for p := 0; p < blockDataLen; p++ {
		blockData[p] = seed
	}

	return longtaillib.CreateStoredBlock(
		uint64(seed)+21412151,
		997,
		2,
		chunkHashes,
		chunkSizes,
		blockData,
		false)
}

func generateUniqueStoredBlock(t *testing.T, seed uint8) (longtaillib.Longtail_StoredBlock, int) {
	chunkHashes := []uint64{uint64(seed)<<8 + 1, uint64(seed)<<8 + 2, uint64(seed)<<8 + 3}
	chunkSizes := []uint32{uint32(seed)<<8 + 10, uint32(seed)<<8 + 20, uint32(seed)<<8 + 30}

	blockDataLen := (int)(chunkSizes[0] + chunkSizes[1] + chunkSizes[2])
	blockData := make([]uint8, blockDataLen)
	for p := 0; p < blockDataLen; p++ {
		blockData[p] = seed
	}

	return longtaillib.CreateStoredBlock(
		uint64(seed)<<16+21412151,
		997,
		2,
		chunkHashes,
		chunkSizes,
		blockData,
		false)
}

func storeBlockFromSeed(t *testing.T, storeAPI longtaillib.Longtail_BlockStoreAPI, seed uint8) (longtaillib.Longtail_StoredBlock, int) {
	storedBlock, errno := generateStoredBlock(t, seed)
	if errno != 0 {
		return longtaillib.Longtail_StoredBlock{}, errno
	}

	p := &putStoredBlockCompletionAPI{}
	p.wg.Add(1)
	errno = storeAPI.PutStoredBlock(storedBlock, longtaillib.CreateAsyncPutStoredBlockAPI(p))
	if errno != 0 {
		p.wg.Done()
		storedBlock.Dispose()
		return longtaillib.Longtail_StoredBlock{}, errno
	}
	p.wg.Wait()

	return storedBlock, p.err
}

func fetchBlockFromStore(t *testing.T, storeAPI longtaillib.Longtail_BlockStoreAPI, blockHash uint64) (longtaillib.Longtail_StoredBlock, int) {
	g := &getStoredBlockCompletionAPI{}
	g.wg.Add(1)
	errno := storeAPI.GetStoredBlock(blockHash, longtaillib.CreateAsyncGetStoredBlockAPI(g))
	if errno != 0 {
		g.wg.Done()
		return longtaillib.Longtail_StoredBlock{}, errno
	}
	g.wg.Wait()

	return g.storedBlock, g.err
}

func validateBlockFromSeed(t *testing.T, seed uint8, storedBlock longtaillib.Longtail_StoredBlock) {
	if !storedBlock.IsValid() {
		t.Errorf("validateBlockFromSeed() g.err %t != %t", storedBlock.IsValid(), true)
	}

	storedBlockIndex := storedBlock.GetBlockIndex()

	if storedBlockIndex.GetBlockHash() != uint64(seed)+21412151 {
		t.Errorf("validateBlockFromSeed() storedBlockIndex.GetBlockHash() %d != %d", storedBlockIndex.GetBlockHash(), uint64(seed)+21412151)
	}

	if storedBlockIndex.GetChunkCount() != 3 {
		t.Errorf("validateBlockFromSeed() storedBlockIndex.GetChunkCount() %d != %d", storedBlockIndex.GetChunkCount(), 3)
	}
}

func TestCreateStoreAndClient(t *testing.T) {
	blobStore, err := NewTestBlobStore("the_path")
	if err != nil {
		t.Errorf("TestCreateStoreAndClient() NewTestBlobStore() %v != %v", err, nil)
	}
	client, err := blobStore.NewClient(context.Background())
	if err != nil {
		t.Errorf("TestCreateStoreAndClient() blobStore.NewClient(context.Background()) %v != %v", err, nil)
	}
	defer client.Close()
}

func TestListObjectsInEmptyStore(t *testing.T) {
	blobStore, _ := NewTestBlobStore("the_path")
	client, _ := blobStore.NewClient(context.Background())
	defer client.Close()
	objects, err := client.GetObjects("")
	if err != nil {
		t.Errorf("TestListObjectsInEmptyStore() client.GetObjects(\"\")) %v != %v", err, nil)
	}
	if len(objects) != 0 {
		t.Errorf("TestListObjectsInEmptyStore() client.GetObjects(\"\")) %d != %d", len(objects), 0)
	}
	obj, _ := client.NewObject("should-not-exist")
	data, err := obj.Read()
	if err != nil || data != nil {
		t.Errorf("TestListObjectsInEmptyStore() obj.Read()) %v != %v", fmt.Errorf("testBlobObject object does not exist: should-not-exist"), err)
	}
}

func TestSingleObjectStore(t *testing.T) {
	blobStore, _ := NewTestBlobStore("the_path")
	client, _ := blobStore.NewClient(context.Background())
	defer client.Close()
	obj, err := client.NewObject("my-fine-object.txt")
	if err != nil {
		t.Errorf("TestSingleObjectStore() client.NewObject(\"my-fine-object.txt\")) %v != %v", err, nil)
	}
	if exists, _ := obj.Exists(); exists {
		t.Errorf("TestSingleObjectStore() obj.Exists()) %t != %t", exists, false)
	}
	testContent := "the content of the object"
	ok, err := obj.Write([]byte(testContent))
	if !ok {
		t.Errorf("TestSingleObjectStore() obj.Write([]byte(testContent)) %t != %t", ok, true)
	}
	if err != nil {
		t.Errorf("TestSingleObjectStore() obj.Write([]byte(testContent)) %v != %v", err, nil)
	}
	data, err := obj.Read()
	if err != nil {
		t.Errorf("TestSingleObjectStore() obj.Read()) %v != %v", err, nil)
	}
	dataString := string(data)
	if dataString != testContent {
		t.Errorf("TestSingleObjectStore() string(data)) %s != %s", dataString, testContent)
	}
	err = obj.Delete()
	if err != nil {
		t.Errorf("TestSingleObjectStore() obj.Delete()) %v != %v", err, nil)
	}
}

func TestDeleteObject(t *testing.T) {
	blobStore, _ := NewTestBlobStore("the_path")
	client, _ := blobStore.NewClient(context.Background())
	defer client.Close()
	obj, _ := client.NewObject("my-fine-object.txt")
	testContent := "the content of the object"
	_, _ = obj.Write([]byte(testContent))
	obj.Delete()
	if exists, _ := obj.Exists(); exists {
		t.Errorf("TestSingleObjectStore() obj.Exists()) %t != %t", exists, false)
	}
}

func TestListObjects(t *testing.T) {
	blobStore, _ := NewTestBlobStore("the_path")
	client, _ := blobStore.NewClient(context.Background())
	defer client.Close()
	obj, _ := client.NewObject("my-fine-object1.txt")
	obj.Write([]byte("my-fine-object1.txt"))
	obj, _ = client.NewObject("my-fine-object2.txt")
	obj.Write([]byte("my-fine-object2.txt"))
	obj, _ = client.NewObject("my-fine-object3.txt")
	obj.Write([]byte("my-fine-object3.txt"))
	objects, err := client.GetObjects("")
	if err != nil {
		t.Errorf("TestListObjects() client.GetObjects(\"\")) %v != %v", err, nil)
	}
	if len(objects) != 3 {
		t.Errorf("TestListObjects() client.GetObjects(\"\")) %d != %d", len(objects), 3)
	}
	for _, o := range objects {
		readObj, err := client.NewObject(o.Name)
		if err != nil {
			t.Errorf("TestListObjects() o.client.NewObject(o.Name)) %d != %d", len(objects), 3)
		}
		if readObj == nil {
			t.Errorf("TestListObjects() o.client.NewObject(o.Name)) %v == %v", readObj, nil)
		}
		data, err := readObj.Read()
		if err != nil {
			t.Errorf("TestListObjects() readObj.Read()) %v != %v", err, nil)
		}
		stringData := string(data)
		if stringData != o.Name {
			t.Errorf("TestListObjects() string(data) != o.Name) %s != %s", stringData, o.Name)
		}
	}
}

func TestStoreIndexSync(t *testing.T) {
	//t.Skip()
	blobStore, _ := NewTestBlobStore("the_path")

	blockGenerateCount := 4
	workerCount := 21

	generatedBlockHashes := make(chan uint64, blockGenerateCount*workerCount)

	var wg sync.WaitGroup
	for n := 0; n < workerCount; n++ {
		wg.Add(1)
		seedBase := blockGenerateCount * n
		go func(blockGenerateCount int, seedBase int) {
			client, _ := blobStore.NewClient(context.Background())
			defer client.Close()
			blocks := []longtaillib.Longtail_BlockIndex{}
			for i := 0; i < blockGenerateCount-1; i++ {
				block, _ := generateUniqueStoredBlock(t, uint8(seedBase+i))
				blocks = append(blocks, block.GetBlockIndex())
			}

			storeIndex, _ := longtaillib.CreateStoreIndexFromBlocks(blocks)
			writeStoreIndex(client, storeIndex)
			storeIndex.Dispose()

			newStoreIndex, _ := readStoreIndex(client)
			newStoreIndex.Dispose()

			for i := blockGenerateCount - 1; i < blockGenerateCount; i++ {
				block, _ := generateUniqueStoredBlock(t, uint8(seedBase+i))
				blocks = append(blocks, block.GetBlockIndex())
			}

			storeIndex, _ = longtaillib.CreateStoreIndexFromBlocks(blocks)
			writeStoreIndex(client, storeIndex)
			defer storeIndex.Dispose()

			newStoreIndex, _ = readStoreIndex(client)
			lookup := map[uint64]bool{}
			for _, h := range newStoreIndex.GetBlockHashes() {
				lookup[h] = true
			}

			blockHashes := storeIndex.GetBlockHashes()
			for n := 0; n < blockGenerateCount; n++ {
				h := blockHashes[n]
				generatedBlockHashes <- h
				_, exists := lookup[h]
				if !exists {
					t.Errorf("TestStoreIndexSync() Missing direct block %d", h)
				}
			}

			wg.Done()
		}(blockGenerateCount, seedBase)
	}
	wg.Wait()
	client, _ := blobStore.NewClient(context.Background())
	defer client.Close()
	newStoreIndex, _ := readStoreIndex(client)
	lookup := map[uint64]bool{}
	for _, h := range newStoreIndex.GetBlockHashes() {
		lookup[h] = true
	}
	for n := 0; n < workerCount*blockGenerateCount; n++ {
		h := <-generatedBlockHashes
		_, exists := lookup[h]
		if !exists {
			t.Errorf("TestStoreIndexSync() Missing block %d", h)
		}
	}
}
