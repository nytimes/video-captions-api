package service

import (
	"context"
	"fmt"
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
	writer.ObjectAttrs.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}

	if _, err := writer.Write(data); err != nil {
		gs.logger.WithError(err).Error("there was some error writing file")
		return "", err
	}
	if err := writer.Close(); err != nil {
		gs.logger.WithError(err).Error("error closing writer ")
		return "", err
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", gs.bucketName, objectFullName), nil
}
