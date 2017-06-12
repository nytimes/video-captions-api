package database

import (
	"errors"
	"sync"

	"github.com/NYTimes/video-captions-api/providers"
)

// MemoryDatabase memory based database implementation for the DB interface
type MemoryDatabase struct {
	mtx  sync.Mutex
	jobs map[string]providers.Job
}

// NewMemoryDatabase creates a MemoryDatabase
func NewMemoryDatabase() *MemoryDatabase {
	return &MemoryDatabase{
		jobs: make(map[string]providers.Job, 0),
	}
}

// StoreJob stores Job in-memory
func (db *MemoryDatabase) StoreJob(job providers.Job) (string, error) {
	if _, err := db.GetJob(job.ID); err == nil {
		return "", errors.New("Job already exists")
	}

	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.jobs[job.ID] = job
	return job.ID, nil
}

// UpdateJob updates Job in-memory
func (db *MemoryDatabase) UpdateJob(id string, job providers.Job) error {
	if _, err := db.GetJob(id); err != nil {
		return errors.New("Job doesn't exist")
	}

	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.jobs[id] = job
	return nil
}

// GetJob returns a Job given its ID
func (db *MemoryDatabase) GetJob(id string) (providers.Job, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	if job, ok := db.jobs[id]; ok {
		return job, nil
	}
	return providers.Job{}, errors.New("Job doesn't exist")
}

// DeleteJob deletes a Job given its ID
func (db *MemoryDatabase) DeleteJob(id string) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	delete(db.jobs, id)
	return nil
}

// GetJobs Returns all Jobs stored
func (db *MemoryDatabase) GetJobs() ([]providers.Job, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	jobList := make([]providers.Job, len(db.jobs))
	for _, job := range db.jobs {
		jobList = append(jobList, job)
	}

	return jobList, nil
}
