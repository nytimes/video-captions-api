package database

import (
	"errors"
	"sync"
)

// MemoryDatabase memory based database implementation for the DB interface
type MemoryDatabase struct {
	mtx  sync.Mutex
	jobs map[string]*Job
}

// NewMemoryDatabase creates a MemoryDatabase
func NewMemoryDatabase() *MemoryDatabase {
	return &MemoryDatabase{
		jobs: make(map[string]*Job, 0),
	}
}

// StoreJob stores Job in-memory
func (db *MemoryDatabase) StoreJob(job *Job) (string, error) {
	if _, err := db.GetJob(job.ID); err == nil {
		return "", errors.New("Job already exists")
	}

	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.jobs[job.ID] = job
	return job.ID, nil
}

// UpdateJob updates Job in-memory
func (db *MemoryDatabase) UpdateJob(id string, job *Job) error {
	if _, err := db.GetJob(id); err != nil {
		return errors.New("Job doesn't exist")
	}

	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.jobs[id] = job
	return nil
}

// GetJob returns a Job given its ID
func (db *MemoryDatabase) GetJob(id string) (*Job, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	if job, ok := db.jobs[id]; ok {
		return job, nil
	}
	return nil, errors.New("Job doesn't exist")
}

// DeleteJob deletes a Job given its ID
func (db *MemoryDatabase) DeleteJob(id string) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	delete(db.jobs, id)
	return nil
}

// GetJobs Returns all Jobs stored for the same ParentID
func (db *MemoryDatabase) GetJobs(parentID string) ([]Job, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	var jobList []Job
	for _, job := range db.jobs {
		if parentID == job.ParentID {
			jobList = append(jobList, *job)
		}
	}

	return jobList, nil
}
