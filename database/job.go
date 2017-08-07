package database

import (
	"time"

	"cloud.google.com/go/datastore"
)

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
	Done           bool           `json:"done"`
}

// JobOutput output associated with a Job
type JobOutput struct {
	URL      string `json:"url"`
	Type     string `json:"type"`
	Filename string `json:"filename"`
}

// ProviderParams is a set of parameters for providers
type ProviderParams map[string]string

// ByCreatedAt implements sort.Interface for []Job by CreatedAt field.
type ByCreatedAt []Job

func (b ByCreatedAt) Len() int { return len(b) }

func (b ByCreatedAt) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b ByCreatedAt) Less(i, j int) bool { return b[i].CreatedAt.Before(b[j].CreatedAt) }

// UpdateStatus update Job status and mark as done if needed
func (j *Job) UpdateStatus(status string) bool {
	if j.Status == status {
		return false
	}
	if status == "error" && !j.Done {
		j.Done = true
	}
	j.Status = status
	return true
}

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
