package providers

import "time"

// Job representation of a captions job
type Job struct {
	ID       string `json:"id"`
	MediaURL string `json:"media_url"`
	Status   string `json:"status"`
	// this should be in the ProviderParams
	ProviderID string `json:"provider_id"`
	Provider   string `json:"provider"`
	//  Datastore doesnt support  maps by default
	ProviderParams map[string]string `datastore:"-" json:"provider_params"`
	CreatedAt      time.Time         `json:"created_at"`
}
