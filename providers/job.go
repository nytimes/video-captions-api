package providers

import "time"

type Job struct {
	ID       string `json:"id"`
	MediaURL string `json:"media_url"`
	Status   string `json:"status"`
	// this should be in the ProviderParams
	ProviderID string `json:"provider_id"`
	Provider   string `json:"provider"`
	//  Datastore doesnt support  maps by default
	//	ProviderMetadata map[string]string `json:"provider_params"`
	CreatedAt time.Time `json:"created_at"`
}
