package providers

// Provider is the interface that transcription/captions providers must implement
type Provider interface {
	DispatchJob(*Job) error
	GetJob(id string) (*Job, error)
	GetName() string
}
