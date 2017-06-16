package providers

import (
	"net/url"
	"strconv"

	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/threeplay"
	captionsConfig "github.com/NYTimes/video-captions-api/config"
	log "github.com/Sirupsen/logrus"
)

const providerName string = "3play"

// ThreePlayProvider is a 3play client that implements the Provider interface
type ThreePlayProvider struct {
	*threeplay.Client
	logger *log.Logger
}

// ThreePlayConfig holds config necessary to create a ThreePlayProvider
type ThreePlayConfig struct {
	APIKey    string `envconfig:"THREE_PLAY_API_KEY"`
	APISecret string `envconfig:"THREE_PLAY_API_SECRET"`
}

// New3PlayProvider creates a ThreePlayProvider instance
func New3PlayProvider(cfg *ThreePlayConfig, svcCfg *captionsConfig.CaptionsServiceConfig) Provider {
	return &ThreePlayProvider{
		threeplay.NewClient(cfg.APIKey, cfg.APISecret),
		svcCfg.Logger,
	}
}

// Load3PlayConfigFromEnv loads 3play API Key/Secret from environment
func Load3PlayConfigFromEnv() ThreePlayConfig {
	var providerConfig ThreePlayConfig
	config.LoadEnvConfig(&providerConfig)
	return providerConfig
}

// GetName returns provider name
func (c *ThreePlayProvider) GetName() string {
	return providerName
}

// GetJob returns a 3play file
func (c *ThreePlayProvider) GetJob(id string) (*Job, error) {
	i, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	file, err := c.GetFile(uint(i))
	if err != nil {
		return nil, err
	}

	job := &Job{
		ID:       strconv.FormatUint(uint64(file.ID), 10),
		Status:   file.State,
		Provider: providerName,
	}
	return job, nil
}

// DispatchJob sends a video file to 3play for transcription and captions generation
func (c *ThreePlayProvider) DispatchJob(job Job) (Job, error) {
	jobLogger := c.logger.WithFields(log.Fields{"JobID": job.ID, "Provider": job.Provider})
	jobLogger.Info("Dispatching job to 3Play")
	// TODO: we need to parse job.ProviderParams to query
	query, _ := url.ParseQuery("for_asr=1")

	fileID, err := c.UploadFileFromURL(job.MediaURL, query)

	if err != nil {
		jobLogger.Error("Failed to dispatch job to 3Play", err)
		return Job{}, err
	}

	job.ProviderID = fileID

	return job, nil
}
