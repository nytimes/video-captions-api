package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
)

type Storage interface {
	Store(data []byte, filename string) (string, error)
}

// GCSStorage implements the Storage interface and stores files on GCS
type GCSStorage struct {
	*storage.Client
	bucket *storage.BucketHandle
}

func NewGCSStorage(bucketName string) (Storage, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	bucket := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	return &GCSStorage{client, bucket}, nil
}

func (gs *GCSStorage) Store(data []byte, filename string) (string, error) {
	ctx := context.Background()
	year, month, day := time.Now().Date()
	obj := gs.bucket.Object(path.Join(strconv.Itoa(year), strconv.Itoa(int(month)), strconv.Itoa(day), filename))
	writer := obj.NewWriter(ctx)
	writer.ContentType = "text/plain"
	writer.ACL = []storage.ACLRule{{storage.AllUsers, storage.RoleReader}}
	if _, err := writer.Write(data); err != nil {
		fmt.Println("there was some error writing file", err)
		return "", nil
	}
	if err := writer.Close(); err != nil {
		fmt.Println("error closing writer ", err)
		return "", nil
	}
	attrs := writer.Attrs()
	fmt.Println("done ", attrs)
	return attrs.MediaLink, nil
}

// FileStore implements the Storage interface and stores files on disk
type FileStorage struct {
	baseDir string
}

// Store stores file on disk
func (s FileStorage) Store(data []byte, filename string) (string, error) {
	dest := fmt.Sprintf("%s/%s", s.baseDir, filename)
	err := ioutil.WriteFile(dest, data, 0644)
	return dest, err
}

// NewFileStorage creates a FileStorage with the specifed base directory
func NewFileStorage(baseDir string) Storage {
	return &FileStorage{baseDir}
}
