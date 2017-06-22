package providers

import (
	"time"

	"cloud.google.com/go/datastore"
)

// Job representation of a captions job
type Job struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	MediaURL string `json:"media_url"`
	Status   string `json:"status"`
	// this should be in the ProviderParams
	ProviderID string `json:"provider_id"`
	Provider   string `json:"provider"`
	//  Datastore doesnt support  maps by default
	ProviderParams ProviderParams `json:"provider_params"`
	CreatedAt      time.Time      `json:"created_at"`
}

// ProviderParams is a set of parameters for providers
type ProviderParams map[string]string

// Load makes ProviderParams implement datastore.PropertyLoadSaver interface
func (p *ProviderParams) Load(ps []datastore.Property) error {
	if *p == nil {
		*p = make(ProviderParams)
	}
	for _, v := range ps {
		(*p)[v.Name] = v.Value.(string)
	}
	return nil
}

// Save makes ProviderParams implement datastore.PropertyLoadSaver interface
func (p *ProviderParams) Save() ([]datastore.Property, error) {
	var result []datastore.Property
	for k, v := range *p {
		result = append(result, datastore.Property{
			Name:  k,
			Value: v,
		})
	}
	return result, nil
}
