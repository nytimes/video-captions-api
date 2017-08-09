package providers

import "github.com/NYTimes/video-captions-api/database"

// Provider is the interface that transcription/captions providers must implement
type Provider interface {
	DispatchJob(*database.Job) error
	Download(id string, captionsType string) ([]byte, error)
	GetProviderJob(id string) (*database.ProviderJob, error)
	GetName() string
}
