package providers

import (
	"time"

	"github.com/nu7hatch/gouuid"

	"cloud.google.com/go/datastore"
)

type JobParams struct {
	ParentID       string         `json:"parent_id"`
	MediaURL       string         `json:"media_url"`
	Provider       string         `json:"provider"`
	ProviderParams ProviderParams `json:"provider_params"`
	OutputTypes    []string       `json:"output_types"`
}

// Job representation of a captions job
type Job struct {
	ID             string         `json:"id"`
	ParentID       string         `json:"parent_id"`
	MediaURL       string         `json:"media_url"`
	Status         string         `json:"status"`
	Provider       string         `json:"provider"`
	ProviderParams ProviderParams `json:"provider_params"`
	CreatedAt      time.Time      `json:"created_at"`
	Outputs        []JobOutput    `json:"outputs"`
	Ended          bool           `json:"ended"`
}

type JobOutput struct {
	Url  string `json:"url"`
	Type string `json:"type"`
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

func NewJobFromParams(newJob JobParams) *Job {
	outputs := make([]JobOutput, 0)
	for _, outputType := range newJob.OutputTypes {
		outputs = append(outputs, JobOutput{Type: outputType})
	}
	id, _ := uuid.NewV4()
	return &Job{
		ID:       id.String(),
		ParentID: newJob.ParentID,
		//TODO: put all possible status under a type/consts so we dont use strings everywhere
		Status:         "processing",
		MediaURL:       newJob.MediaURL,
		Provider:       newJob.Provider,
		ProviderParams: newJob.ProviderParams,
		CreatedAt:      time.Now(),
		Outputs:        outputs,
		Ended:          false,
	}
}
