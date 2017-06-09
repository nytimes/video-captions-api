package database

import "github.com/NYTimes/video-captions-api/providers"

type DB interface {
	StoreJob(providers.Job) (string, error)
	UpdateJob(string, providers.Job) error
	GetJob(string) (providers.Job, error)
	DeleteJob(string) error
	GetJobs() ([]providers.Job, error)
}
