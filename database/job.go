package database

import (
	"time"
	"fmt"

	"cloud.google.com/go/datastore"
	uuid "github.com/nu7hatch/gouuid"
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

type JobOutput struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

func (output JobOutput) Name() string {
  id, _ := uuid.NewV4()
  name := id.String()[:8]
  return fmt.Sprintf("%s.%s", name, output.Type)
}

// ProviderParams is a set of parameters for providers
type ProviderParams map[string]string

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
