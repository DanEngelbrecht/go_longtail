package longtailstorelib

import (
	"log"
	"net/url"
	"sync"
	"testing"

	"github.com/DanEngelbrecht/golongtail/longtaillib"
	"golang.org/x/net/context"
)

func TestGCSBlobStore(t *testing.T) {
	// This test uses hardcoded paths in gcs and is disabled
	t.Skip()

	u, err := url.Parse("gs://longtail-test/test-gcs-blob-store")
	if err != nil {
		t.Errorf("url.Parse() err == %q", err)
	}
	blobStore, err := NewGCSBlobStore(u)
	if err != nil {
		t.Errorf("NewGCSBlobStore() err == %q", err)
	}
	client, err := blobStore.NewClient(context.Background())
	if err != nil {
		t.Errorf("blobStore.NewClient() err == %q", err)
	}
	object, err := client.NewObject("test.txt")
	if err != nil {
		t.Errorf("client.NewObject() err == %q", err)
	}
	testData := []byte("apa")
	ok, err := object.Write(testData)
	if !ok {
		t.Errorf("object.Write() ok != true")
	}
	if err != nil {
		t.Errorf("object.Write() err == %q", err)
	}

	blobs, err := client.GetObjects("")
	if err != nil {
		t.Errorf("client.GetObjects(\"\") err == %q", err)
	}
	if blobs[0].Name != "test.txt" {
		t.Errorf("blobs[0].Name %s != %s", blobs[0].Name, "test.txt")
	}
	data, err := object.Read()
	if len(data) != 3 {
		t.Errorf("len(data) %d != %d", len(data), 3)
	}
	for i, d := range data {
		if d != testData[i] {
			t.Errorf("%d != testData[%d]", int(d), int(testData[i]))
		}
	}

	object, _ = client.NewObject("path/first.txt")
	_, _ = object.Write([]byte("dog"))

	object, _ = client.NewObject("path/second.txt")
	_, _ = object.Write([]byte("cat"))

	objects, _ := client.GetObjects("")
	if len(objects) != 3 {
		t.Errorf("len(objects) %d != 3", len(objects))
	}

	objects, _ = client.GetObjects("path/")
	if len(objects) != 2 {
		t.Errorf("len(objects) %d != 2", len(objects))
	}
}

func TestGCSStoreIndexSync(t *testing.T) {
	// This test uses hardcoded paths in S3 and is disabled
	t.Skip()

	u, err := url.Parse("gs://longtail-test-de/test-gcs-blob-store-sync")
	if err != nil {
		t.Errorf("url.Parse() err == %q", err)
	}

	blobStore, err := NewGCSBlobStore(u)
	if err != nil {
		log.Fatalf("%v", err)
	}

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
