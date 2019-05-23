package providers

import (
	"fmt"

	captionsConfig "github.com/NYTimes/video-captions-api/config"
	"github.com/NYTimes/video-captions-api/database"
	log "github.com/sirupsen/logrus"
)

// UploadProvider in a GCP client wrapper that implements the Provider interface
type UploadProvider struct {
	logger *log.Logger
	DB     database.DB
}

// NewUploadProvider initializes the GCP provider.
func NewUploadProvider(svcCfg *captionsConfig.CaptionsServiceConfig, db database.DB) Provider {
	return &UploadProvider{
		svcCfg.Logger,
		db,
	}
}

// GetName returns the name of the upload provider - GCP.
func (c *UploadProvider) GetName() string {
	return "upload"
}

// Download returns the uploaded caption file
func (c *UploadProvider) Download(id string, captionsType string) ([]byte, error) {
	job, err := c.DB.GetJob(id)
	if err != nil {
		return nil, fmt.Errorf("could not find job in DB")
	}
	return job.CaptionFile.File, nil
}

// GetProviderJob returns the provider's job parameters.
func (c *UploadProvider) GetProviderJob(id string) (*database.ProviderJob, error) {
	job, err := c.DB.GetJob(id)
	if err != nil {
		return nil, fmt.Errorf("could not find job in DB")
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
func (c *UploadProvider) DispatchJob(job *database.Job) error {
	job.Status = "delivered"
	job.ProviderParams = map[string]string{
		"ProviderID": job.ID,
		"status":     "delivered",
		"details":    "Version 1",
	}
	return nil
}

func (c *UploadProvider) CancelJob(id string) (bool, error) {
	return false, nil
}
