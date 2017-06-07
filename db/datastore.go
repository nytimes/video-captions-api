package database

import (
	"context"
	"strconv"

	"cloud.google.com/go/datastore"
)

const (
	entityKind      string = "Jobs"
	entityNamespace string = "captions-jobs"
)

// DatastoreClient is a datastore client that implements DB interface
type DatastoreClient struct {
	client    *datastore.Client
	kind      string
	namespace string
}

// NewDatastoreClient returns a DatastoreClient
func NewDatastoreClient(projectID string) (*DatastoreClient, error) {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &DatastoreClient{
		client,
		entityKind,
		entityNamespace,
	}, nil

}

// StoreJob stores a job
func (c *DatastoreClient) StoreJob(job Job) (string, error) {
	ctx := context.Background()
	key := makeKey(c.kind, c.namespace)
	newKey, err := c.client.Put(ctx, key, &job)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(newKey.ID, 10), nil
}

// GetJob retrieves a job from database
func (c *DatastoreClient) GetJob(idStr string) (Job, error) {
	result := Job{}
	ctx := context.Background()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return result, err
	}

	key := datastore.IDKey(c.kind, id, nil)
	key.Namespace = c.namespace

	err := c.client.Get(ctx, key, &result)
	if err != nil {
		return result, err
	}

	return result, nil

}

// GetJobs retrieves all jobs in database
func (c *DatastoreClient) GetJobs() ([]Job, error) {
	query := datastore.NewQuery(c.kind).Namespace(c.namespace)
	var jobs []Job
	ctx := context.Background()
	_, err := c.client.GetAll(ctx, query, &jobs)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// UpdateJob updates a job
func (c *DatastoreClient) UpdateJob(idStr string, job Job) error {
	_, err := c.GetJob(idStr)
	if err != nil {
		return err
	}

	ctx := context.Background()
	key := makeKey(c.kind, c.namespace)
	key.ID, err = strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}
	_, err = c.client.Put(ctx, key, &job)
	return err
}

// DeleteJob deletes a job from database
func (c *DatastoreClient) DeleteJob(idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}

	ctx := context.Background()
	key := datastore.IDKey(c.kind, id, nil)
	key.Namespace = c.namespace
	return c.client.Delete(ctx, key)
}

func makeKey(kind, namespace string) *datastore.Key {
	key := datastore.IncompleteKey(kind, nil)
	key.Namespace = namespace
	return key
}
