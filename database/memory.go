package database

import (
	"errors"
	"sync"

	"github.com/NYTimes/video-captions-api/providers"
)

type MemoryDatabse struct {
	mtx  sync.Mutex
	jobs map[string]providers.Job
}

func NewMemoryDatabase() *MemoryDatabse {
	return &MemoryDatabse{
		jobs: make(map[string]providers.Job, 0),
	}
}

func (db *MemoryDatabse) StoreJob(job providers.Job) (string, error) {
	if _, err := db.GetJob(job.ID); err == nil {
		return "", errors.New("Job already exists")
	}

	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.jobs[job.ID] = job
	return job.ID, nil
}

func (db *MemoryDatabse) UpdateJob(id string, job providers.Job) error {
	if _, err := db.GetJob(id); err != nil {
		return errors.New("Job doesn't exist")
	}

	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.jobs[id] = job
	return nil
}
func (db *MemoryDatabse) GetJob(id string) (providers.Job, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	if job, ok := db.jobs[id]; ok {
		return job, nil
	}
	return providers.Job{}, errors.New("Job doesn't exist")
}
func (db *MemoryDatabse) DeleteJob(id string) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	delete(db.jobs, id)
	return nil
}
func (db *MemoryDatabse) GetJobs() ([]providers.Job, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	jobList := make([]providers.Job, len(db.jobs))
	for _, job := range db.jobs {
		jobList = append(jobList, job)
	}

	return jobList, nil
}
