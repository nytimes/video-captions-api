package database

import (
	"errors"

	"golang.org/x/net/context"

	"github.com/NYTimes/video-captions-api/providers"

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
func (d *DatastoreDatabase) StoreJob(job *providers.Job) (string, error) {
	if _, err := d.GetJob(job.ID); err == nil {
		return "", errors.New("Job already exists")
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
func (d *DatastoreDatabase) GetJob(id string) (*providers.Job, error) {
	result := &providers.Job{}
	ctx := context.Background()
	key := newNameKeyWithNamespace(d.kind, id, d.namespace)
	err := d.client.Get(ctx, key, result)
	if err == datastore.ErrNoSuchEntity {
		return nil, errors.New("Job not found")
	}
	if err != nil {
		return nil, errors.New("Unkown error from Datastore")
	}
	return result, nil
}

// GetJobs retrieves all jobs in database
func (d *DatastoreDatabase) GetJobs(parentID string) ([]providers.Job, error) {
	var jobs []providers.Job
	ctx := context.Background()
	query := datastore.NewQuery(d.kind).Namespace(d.namespace).Filter("ParentID =", parentID)
	_, err := d.client.GetAll(ctx, query, &jobs)
	if err != nil {
		return nil, errors.New("Unkown error from Datastore")
	}
	if len(jobs) == 0 {
		return nil, errors.New("No Jobs found for this ParentID")
	}
	return jobs, nil
}

// UpdateJob updates a job
func (d *DatastoreDatabase) UpdateJob(id string, job *providers.Job) error {
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

func newNameKeyWithNamespace(kind, name, namespace string) *datastore.Key {
	key := datastore.NameKey(kind, name, nil)
	key.Namespace = namespace
	return key
}
