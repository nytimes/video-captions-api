package database

import "errors"

var (
	// ErrNoJobs indicates that no jobs can be found for a given parent ID.
	ErrNoJobs = errors.New("no jobs found with this parent ID")

	// ErrJobNotFound indicates that no job can be found for a given job ID.
	ErrJobNotFound = errors.New("job not found")
)

// DB interface for database interactions
type DB interface {
	StoreJob(*Job) (string, error)
	UpdateJob(string, *Job) error
	GetJob(string) (*Job, error)
	DeleteJob(string) error
	GetJobs(string) ([]Job, error)
}
