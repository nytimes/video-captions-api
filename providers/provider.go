package providers

import "github.com/NYTimes/video-captions-api/database"

// Provider is the interface that transcription/captions providers must implement
type Provider interface {
	DispatchJob(*database.Job) error
	GetJob(id string) (*database.Job, error)
	GetName() string
}
