package database

import "time"

type DB interface {
	StoreJob(Job) (string, error)
	UpdateJob(string, Job) error
	GetJob(string) (Job, error)
	DeleteJob(string) error
	GetJobs() ([]Job, error)
}

type Job struct {
	Media_url string
	Status    string
	ScoopID   string
	CreatedAt time.Time
}
