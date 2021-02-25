package longtailstorelib

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

type fsBlobStore struct {
	prefix string
}

type fsBlobClient struct {
	store *fsBlobStore
}

type fsBlobObject struct {
	client *fsBlobClient
	path   string
}

// NewFSBlobStore ...
func NewFSBlobStore(prefix string) (BlobStore, error) {
	s := &fsBlobStore{prefix: prefix}
	return s, nil
}

func (blobStore *fsBlobStore) NewClient(ctx context.Context) (BlobClient, error) {
	return &fsBlobClient{store: blobStore}, nil
}

func (blobStore *fsBlobStore) String() string {
	return "fsstore"
}

func (blobClient *fsBlobClient) NewObject(filepath string) (BlobObject, error) {
	fsPath := path.Join(blobClient.store.prefix, filepath)
	return &fsBlobObject{client: blobClient, path: fsPath}, nil
}

func (blobClient *fsBlobClient) GetObjects(pathPrefix string) ([]BlobProperties, error) {
	return make([]BlobProperties, 0), nil
}

func (blobClient *fsBlobClient) Close() {
}

func (blobClient *fsBlobClient) String() string {
	return "fsstore"
}

func (blobObject *fsBlobObject) Exists() (bool, error) {
	_, err := os.Stat(blobObject.path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (blobObject *fsBlobObject) Read() ([]byte, error) {
	data, err := ioutil.ReadFile(blobObject.path)
	if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (blobObject *fsBlobObject) Write(data []byte) (bool, error) {
	err := os.MkdirAll(filepath.Dir(blobObject.path), os.ModePerm)
	if err != nil {
		return false, err
	}
	err = ioutil.WriteFile(blobObject.path, data, 0644)
	if err != nil {
		return false, err
	}
	return true, err
}

func (blobObject *fsBlobObject) Delete() error {
	return os.Remove(blobObject.path)
}
