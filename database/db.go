package database

// DB interface for database interactions
type DB interface {
	StoreJob(*Job) (string, error)
	UpdateJob(string, *Job) error
	GetJob(string) (*Job, error)
	DeleteJob(string) error
	GetJobs(string) ([]Job, error)
}
