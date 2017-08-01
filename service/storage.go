package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"

	"cloud.google.com/go/storage"
)

// Storage inferce for captions storage
type Storage interface {
	Store(data []byte, filename string) (string, error)
}

// GCSStorage implements the Storage interface and stores files on GCS
type GCSStorage struct {
	bucketHandle *storage.BucketHandle
	bucketName   string
	logger       *log.Logger
}

// NewGCSStorage creates a GCSStorage instance
func NewGCSStorage(bucketName string, logger *log.Logger) (Storage, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	bucket := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	return &GCSStorage{bucket, bucketName, logger}, nil
}

// Store implements Storage interface for GCSStorage
func (gs *GCSStorage) Store(data []byte, filename string) (string, error) {
	ctx := context.Background()
	year, month, day := time.Now().Date()
	objectFullName := fmt.Sprintf("%s/%s/%s/%s", strconv.Itoa(year), strconv.Itoa(int(month)), strconv.Itoa(day), filename)
	obj := gs.bucketHandle.Object(objectFullName)
	writer := obj.NewWriter(ctx)
	writer.ContentType = "text/plain"
	if _, err := writer.Write(data); err != nil {
		gs.logger.WithError(err).Info("there was some error writing file")
		return "", err
	}
	if err := writer.Close(); err != nil {
		gs.logger.WithError(err).Info("error closing writer ")
		return "", err
	}

	return fmt.Sprintf("gs://%s/%s", gs.bucketName, objectFullName), nil
}

// FileStorage implements the Storage interface and stores files on disk
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
