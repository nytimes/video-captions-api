package providers

type Provider interface {
	DispatchJob(*Job) (*Job, error)
	GetJob(id string) (*Job, error)
}
