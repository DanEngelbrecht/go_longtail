package longtailstorelib

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
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
	defer blobClient.store.blobsMutex.RUnlock()
	properties := make([]BlobProperties, 0)
	i := 0
	for key, blob := range blobClient.store.blobs {
		if strings.HasPrefix(key, pathPrefix) {
			properties = append(properties, BlobProperties{Name: key, Size: int64(len(blob.data))})
		}
		i++
	}
	return properties, nil
}

func (blobClient *testBlobClient) Close() {
}

func (blobClient *testBlobClient) String() string {
	return "teststore"
}

func (blobObject *testBlobObject) Exists() (bool, error) {
	blobObject.client.store.blobsMutex.RLock()
	defer blobObject.client.store.blobsMutex.RUnlock()
	_, exists := blobObject.client.store.blobs[blobObject.path]
	return exists, nil
}

func (blobObject *testBlobObject) Read() ([]byte, error) {
	blobObject.client.store.blobsMutex.RLock()
	defer blobObject.client.store.blobsMutex.RUnlock()
	blob, exists := blobObject.client.store.blobs[blobObject.path]
	if !exists {
		return nil, fmt.Errorf("testBlobObject object does not exist: %s", blobObject.path)
	}
	return blob.data, nil
}

func (blobObject *testBlobObject) Write(data []byte) (bool, error) {
	blobObject.client.store.blobsMutex.Lock()
	defer blobObject.client.store.blobsMutex.Unlock()

	blob, exists := blobObject.client.store.blobs[blobObject.path]

	if !exists {
		blob = &testBlob{path: blobObject.path, data: data}
		blobObject.client.store.blobs[blobObject.path] = blob
		return true, nil
	}

	blob.data = data
	return true, nil
}

func (blobObject *testBlobObject) Delete() error {
	blobObject.client.store.blobsMutex.Lock()
	defer blobObject.client.store.blobsMutex.Unlock()

	delete(blobObject.client.store.blobs, blobObject.path)
	return nil
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
	if err == nil {
		t.Errorf("TestListObjectsInEmptyStore() obj.Read()) %v != %v", fmt.Errorf("testBlobObject object does not exist: should-not-exist"), err)
	}
	if data != nil {
		t.Errorf("TestListObjectsInEmptyStore() obj.Read()) %v != %v", nil, data)
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
