package providers

// Provider is the interface that transcription/captions providers must implement
type Provider interface {
	DispatchJob(Job) (Job, error)
	GetJob(id string) (*Job, error)
	GetName() string
	GetOptions() []ProviderOption
}

//ProviderOption is a provider option
type ProviderOption struct {
	Key         string   `json:"key"`
	Value       []string `json:"value"`
	Description string   `json:"description"`
}
