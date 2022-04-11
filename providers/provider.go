package providers

import "github.com/nytimes/video-captions-api/database"

// Provider is the interface that transcription/captions providers must implement
type Provider interface {
	DispatchJob(*database.Job) error
	Download(*database.Job, string) ([]byte, error)
	GetProviderJob(*database.Job) (*database.ProviderJob, error)
	GetName() string
	CancelJob(*database.Job) (bool, error)
}
