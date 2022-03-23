package providers

import (
	"bytes"
	"fmt"
	"path/filepath"

	captionsConfig "github.com/nytimes/video-captions-api/config"
	"github.com/nytimes/video-captions-api/database"
	"github.com/nytimes/video-captions-api/vtt"
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
func (c *UploadProvider) Download(job *database.Job, captionsType string) ([]byte, error) {
	job, err := c.DB.GetJob(job.GetProviderID())
	if err != nil {
		return nil, fmt.Errorf("could not find job in DB")
	}
	return job.CaptionFile.File, nil
}

// GetProviderJob returns the provider's job parameters.
func (c *UploadProvider) GetProviderJob(job *database.Job) (*database.ProviderJob, error) {
	job, err := c.DB.GetJob(job.GetProviderID())
	if err != nil {
		return nil, fmt.Errorf("could not find job in DB")
	}
	providerJob := &database.ProviderJob{
		ID:      job.GetProviderID(),
		Status:  job.ProviderParams["status"],
		Details: job.ProviderParams["details"],
	}
	return providerJob, nil
}

// DispatchJob sets the status of the upload job as delivered so
// that a call to check the job status uploads it to the cloud.
func (c *UploadProvider) DispatchJob(job *database.Job) error {
	err := c.validateCaptionFile(&job.CaptionFile)

	if err != nil {
		return err
	}

	job.Status = "delivered"
	job.ProviderParams = map[string]string{
		"ProviderID": job.ID,
		"status":     "delivered",
		"details":    "Version 1",
	}
	return nil
}

func (c *UploadProvider) CancelJob(job *database.Job) (bool, error) {
	return false, nil
}

// validateCaptionFile checks the contents of the uploaded file to
// ensure it's a valid captions fle. It uses extension to determine
// which type of file check to perform. Currently only .vtt validation
// is supported.
func (c *UploadProvider) validateCaptionFile(file *database.UploadedFile) error {
	ext := filepath.Ext(file.Name)

	if ext == ".vtt" {
		return vtt.Validate(bytes.NewReader(file.File))
	}

	return nil
}
