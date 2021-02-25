package longtailstorelib

import (
	"testing"

	"golang.org/x/net/context"
)

func TestFSBlobStore(t *testing.T) {
	// This test uses hardcoded paths and is disabled
	t.Skip()

	blobStore, err := NewFSBlobStore("C:\\Temp\\fsblobstore")
	if err != nil {
		t.Errorf("NewFSBlobStore() err == %q", err)
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
}
