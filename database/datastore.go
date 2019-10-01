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

// DatastoreClient is a datastore interface with operations used by the captions API
type DatastoreClient interface {
	Put(context.Context, *datastore.Key, interface{}) (*datastore.Key, error)
	Get(context.Context, *datastore.Key, interface{}) error
	Delete(context.Context, *datastore.Key) error
	GetAll(context.Context, *datastore.Query, interface{}) ([]*datastore.Key, error)
}

// DatastoreDatabase is a datastore client that implements DB interface
type DatastoreDatabase struct {
	client    DatastoreClient
	kind      string
	namespace string
}

// NewDatastoreDatabase returns a DatastoreDatabase
func NewDatastoreDatabase(projectID string) (*DatastoreDatabase, error) {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &DatastoreDatabase{
		client,
		entityKind,
		entityNamespace,
	}, nil
}

// StoreJob stores a job
func (d *DatastoreDatabase) StoreJob(job *Job) (string, error) {
	if _, err := d.GetJob(job.ID); err == nil {
		return "", errors.New("job already exists")
	}

	ctx := context.Background()
	key := newNameKeyWithNamespace(d.kind, job.ID, d.namespace)
	_, err := d.client.Put(ctx, key, job)
	if err != nil {
		return "", err
	}

	return job.ID, nil
}

// GetJob retrieves a job from database
func (d *DatastoreDatabase) GetJob(id string) (*Job, error) {
	result := &Job{}
	ctx := context.Background()
	key := newNameKeyWithNamespace(d.kind, id, d.namespace)
	err := d.client.Get(ctx, key, result)
	if err == datastore.ErrNoSuchEntity {
		return nil, ErrJobNotFound
	}
	if err != nil {
		return nil, errors.New("unknown error from Datastore")
	}
	return result, nil
}

// GetJobs retrieves all jobs in database
func (d *DatastoreDatabase) GetJobs(parentID string) ([]Job, error) {
	var jobs []Job
	ctx := context.Background()
	query := datastore.NewQuery(d.kind).Namespace(d.namespace).Filter("ParentID =", parentID)
	_, err := d.client.GetAll(ctx, query, &jobs)
	if err != nil {
		return nil, errors.New("unknown error from Datastore")
	}
	if len(jobs) == 0 {
		return nil, ErrNoJobs
	}
	return jobs, nil
}

// UpdateJob updates a job
func (d *DatastoreDatabase) UpdateJob(id string, job *Job) error {
	_, err := d.GetJob(id)
	if err != nil {
		return err
	}

	ctx := context.Background()
	key := newNameKeyWithNamespace(d.kind, id, d.namespace)
	_, err = d.client.Put(ctx, key, job)
	return err
}

// DeleteJob deletes a job from database
func (d *DatastoreDatabase) DeleteJob(id string) error {
	ctx := context.Background()
	key := newNameKeyWithNamespace(d.kind, id, d.namespace)
	return d.client.Delete(ctx, key)
}

// GetJobByProviderID returns a job associated with a given provider ID
func (d *DatastoreDatabase) GetJobByProviderID(providerID string) (*Job, error) {
	ctx := context.Background()
	query := datastore.NewQuery(d.kind).Namespace(d.namespace).Filter("ProviderParams.ProviderID =", providerID).Limit(1)
	var jobs []Job
	if _, err := d.client.GetAll(ctx, query, &jobs); err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, ErrNoJobs
	}
	return &jobs[0], nil
}

func newNameKeyWithNamespace(kind, name, namespace string) *datastore.Key {
	key := datastore.NameKey(kind, name, nil)
	key.Namespace = namespace
	return key
}
