package database

import (
	"context"
	"errors"

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
	if _, err := c.GetJob(job.ID); err == nil {
		return "", errors.New("Job already exists")
	}

	ctx := context.Background()
	key := NewNameKeyWithNamespace(c.kind, job.ID, c.namespace)
	_, err := c.client.Put(ctx, key, &job)
	if err != nil {
		return "", err
	}

	return job.ID, nil
}

// GetJob retrieves a job from database
func (c *DatastoreClient) GetJob(id string) (Job, error) {
	result := Job{}
	ctx := context.Background()
	key := NewNameKeyWithNamespace(c.kind, id, c.namespace)
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
func (c *DatastoreClient) UpdateJob(id string, job Job) error {
	_, err := c.GetJob(id)
	if err != nil {
		return err
	}

	ctx := context.Background()
	key := NewNameKeyWithNamespace(c.kind, id, c.namespace)
	_, err = c.client.Put(ctx, key, &job)
	return err
}

// DeleteJob deletes a job from database
func (c *DatastoreClient) DeleteJob(id string) error {
	ctx := context.Background()
	key := NewNameKeyWithNamespace(c.kind, id, c.namespace)
	return c.client.Delete(ctx, key)
}

func NewNameKeyWithNamespace(kind, name, namespace string) *datastore.Key {
	key := datastore.NameKey(kind, name, nil)
	key.Namespace = namespace
	return key
}
