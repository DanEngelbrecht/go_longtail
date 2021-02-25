package longtailstorelib

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type s3BlobStore struct {
	bucketName string
	prefix     string
}

type s3BlobClient struct {
	ctx    context.Context
	store  *s3BlobStore
	client *s3.Client
}

type s3BlobObject struct {
	ctx    context.Context
	client *s3BlobClient
	path   string
}

// NewS3BlobStore ...
func NewS3BlobStore(u *url.URL) (BlobStore, error) {
	if u.Scheme != "s3" {
		return nil, fmt.Errorf("invalid scheme '%s', expected 'gs'", u.Scheme)
	}
	prefix := u.Path
	if len(u.Path) > 0 {
		prefix = u.Path[1:] // strip initial slash
	}

	if prefix != "" {
		prefix += "/"
	}
	s := &s3BlobStore{bucketName: u.Host, prefix: prefix}
	return s, nil
}

func (blobStore *s3BlobStore) NewClient(ctx context.Context) (BlobClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	client := s3.NewFromConfig(cfg)
	return &s3BlobClient{store: blobStore, ctx: ctx, client: client}, nil
}

func (blobStore *s3BlobStore) String() string {
	return "s3://" + blobStore.bucketName + "/" + blobStore.prefix
}

func (blobClient *s3BlobClient) NewObject(path string) (BlobObject, error) {
	s3Path := blobClient.store.prefix + path
	return &s3BlobObject{
			ctx:    blobClient.ctx,
			client: blobClient,
			path:   s3Path},
		nil
}

func (blobClient *s3BlobClient) GetObjects(pathPrefix string) ([]BlobProperties, error) {
	var items []BlobProperties
	output, err := blobClient.client.ListObjectsV2(blobClient.ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(blobClient.store.bucketName),
		Prefix: aws.String(blobClient.store.prefix + pathPrefix),
	})
	if err != nil {
		return nil, err
	}
	for _, object := range output.Contents {
		itemName := aws.ToString(object.Key)[len(blobClient.store.prefix):]
		items = append(items, BlobProperties{Size: object.Size, Name: itemName})
	}
	return items, nil
}

func (blobClient *s3BlobClient) Close() {
	blobClient.client = nil
}

func (blobClient *s3BlobClient) String() string {
	return blobClient.store.String()
}

func (blobObject *s3BlobObject) Read() ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(blobObject.client.store.bucketName),
		Key:    aws.String(blobObject.path),
	}
	result, err := blobObject.client.client.GetObject(blobObject.client.ctx, input)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, nil
		}
		return nil, err
	}
	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	result.Body.Close()
	return data, nil
}

func (blobObject *s3BlobObject) Exists() (bool, error) {
	input := &s3.GetObjectAclInput{
		Bucket: aws.String(blobObject.client.store.bucketName),
		Key:    aws.String(blobObject.path),
	}
	_, err := blobObject.client.client.GetObjectAcl(blobObject.client.ctx, input)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (blobObject *s3BlobObject) Write(data []byte) (bool, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(blobObject.client.store.bucketName),
		Key:    aws.String(blobObject.path),
		Body:   bytes.NewReader(data),
	}
	_, err := blobObject.client.client.PutObject(blobObject.client.ctx, input)
	if err != nil {
		fmt.Println(err.Error())
		return true, err
	}
	return true, nil
}

func (blobObject *s3BlobObject) Delete() error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(blobObject.client.store.bucketName),
		Key:    aws.String(blobObject.path),
	}
	_, err := blobObject.client.client.DeleteObject(blobObject.client.ctx, input)
	return err
}
