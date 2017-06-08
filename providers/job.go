package providers

type Job struct {
	ID             string                 `json:"id"`
	MediaURL       string                 `json:"media_url"`
	Status         string                 `json:"status"`
	Provider       string                 `json:"provider"`
	ProviderParams map[string]interface{} `json:"provider_params"`
}
