package database

import (
	"errors"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"cloud.google.com/go/datastore"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/stretchr/testify/assert"
)

type datastoreTestClient struct {
	jobs map[string]providers.Job
}

func (c *datastoreTestClient) Put(_ context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	job := *src.(*providers.Job)
	c.jobs[key.Name] = job
	return key, nil
}

func (c *datastoreTestClient) Get(_ context.Context, key *datastore.Key, dst interface{}) error {
	job, ok := c.jobs[key.Name]
	if !ok {
		return errors.New("ErrNoSuchEntity")
	}

	v := reflect.ValueOf(dst)
	v.Elem().Set(reflect.ValueOf(job))
	return nil
}

func (c *datastoreTestClient) Delete(_ context.Context, key *datastore.Key) error {
	delete(c.jobs, key.Name)
	return nil
}

func (c *datastoreTestClient) GetAll(_ context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error) {
	return nil, nil
}

func newTestDB() *DatastoreDatabase {
	return &DatastoreDatabase{
		&datastoreTestClient{
			map[string]providers.Job{},
		},
		"kind",
		"namespace",
	}
}

func TestStoreJob(t *testing.T) {
	assert := assert.New(t)
	db := newTestDB()

	job := providers.Job{
		ID:       "123",
		MediaURL: "https://abc.com/123.mp4",
	}

	assert.Equal(0, len(db.client.(*datastoreTestClient).jobs))
	id, err := db.StoreJob(job)
	assert.Equal(job.ID, id)
	assert.Nil(err)
	assert.Equal(1, len(db.client.(*datastoreTestClient).jobs))

	_, err = db.StoreJob(job)
	assert.EqualError(err, "Job already exists")
}

func TestGetJob(t *testing.T) {
	assert := assert.New(t)
	db := newTestDB()

	job := providers.Job{
		ID:       "123",
		MediaURL: "https://abc.com/123.mp4",
	}

	db.StoreJob(job)
	result, err := db.GetJob("123")
	assert.Equal(job, result)
	assert.Nil(err)

	_, err = db.GetJob("456")
	assert.EqualError(err, "ErrNoSuchEntity")
}

func TestUpdateJob(t *testing.T) {
	assert := assert.New(t)
	db := newTestDB()

	job := providers.Job{
		ID:       "123",
		MediaURL: "https://abc.com/some.mp4",
	}

	newJob := providers.Job{
		ID:       "123",
		MediaURL: "https://abc.com/another.mp4",
	}
	db.StoreJob(job)
	err := db.UpdateJob("123", newJob)
	assert.Nil(err)
	assert.Equal(1, len(db.client.(*datastoreTestClient).jobs))
	afterChange, err := db.GetJob("123")
	assert.Equal(newJob, afterChange)

	err = db.UpdateJob("234", newJob)
	assert.EqualError(err, "ErrNoSuchEntity")
}

func TestDeleteJob(t *testing.T) {
	assert := assert.New(t)
	db := newTestDB()

	job := providers.Job{
		ID:       "123",
		MediaURL: "https://abc.com/123.mp4",
	}

	db.StoreJob(job)
	assert.Equal(1, len(db.client.(*datastoreTestClient).jobs))
	err := db.DeleteJob("123")
	assert.Nil(err)
	assert.Equal(0, len(db.client.(*datastoreTestClient).jobs))
}
