package providers

import (
	"fmt"
	captionsConfig "github.com/NYTimes/video-captions-api/config"
	"github.com/NYTimes/video-captions-api/database"
	log "github.com/Sirupsen/logrus"
)

// GCPProvider in a GCP client wrapper that implements the Provider interface
type GCPProvider struct {
	logger *log.Logger
	DB     database.DB
}

// NewGCPProvider initializes the GCP provider.
func NewGCPProvider(svcCfg *captionsConfig.CaptionsServiceConfig, db database.DB) Provider {
	return &GCPProvider{
		svcCfg.Logger,
		db,
	}
}

// GetName returns the name of the upload provider - GCP.
func (c *GCPProvider) GetName() string {
	return "gcp"
}

// Download returns the uploaded caption file
func (c *GCPProvider) Download(id string, captionsType string) ([]byte, error) {
	job, err := c.DB.GetJob(id)
	if err != nil {
		return nil, fmt.Errorf("Could not find job in DB")
	}
	return job.CaptionFile.File, nil
}

// GetProviderJob returns the provider's job parameters.
func (c *GCPProvider) GetProviderJob(id string) (*database.ProviderJob, error) {
	job, err := c.DB.GetJob(id)
	if err != nil {
		return nil, fmt.Errorf("Could not find job in DB")
	}
	providerJob := &database.ProviderJob{
		ID:      job.ProviderParams["ProviderID"],
		Status:  job.ProviderParams["status"],
		Details: job.ProviderParams["details"],
	}
	return providerJob, nil
}

// DispatchJob sets the status of the upload job as delivered so
// that a call to check the job status uploads it to the cloud.
func (c *GCPProvider) DispatchJob(job *database.Job) error {
	job.Status = "delivered"
	job.ProviderParams = map[string]string{
		"ProviderID": job.ID,
		"status":     "delivered",
		"details":    "Version 1",
	}
	return nil
}
