package providers

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/NYTimes/video-captions-api/config"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/kelseyhightower/envconfig"
	"github.com/nytimes/threeplay/types"
	threeplay "github.com/nytimes/threeplay/v3api"
	log "github.com/sirupsen/logrus"
)

const providerName string = "3play"

// ThreePlayProvider is a 3play client that implements the Provider interface
type ThreePlayProvider struct {
	*threeplay.Client
	logger *log.Logger
	config ThreePlayConfig
}

// ThreePlayConfig holds config necessary to create a ThreePlayProvider
type ThreePlayConfig struct {
	APIKeyByJobType map[string]string `envconfig:"THREE_PLAY_API_KEY"`
}

// New3PlayProvider creates a ThreePlayProvider instance
func New3PlayProvider(cfg *ThreePlayConfig, svcCfg *config.CaptionsServiceConfig) Provider {
	return &ThreePlayProvider{
		threeplay.NewClient(cfg.APIKeyByJobType["captions"]),
		svcCfg.Logger,
		*cfg,
	}
}

// Load3PlayConfigFromEnv loads 3play API Key/Secret from environment
func Load3PlayConfigFromEnv() ThreePlayConfig {
	var providerConfig ThreePlayConfig
	envconfig.Process("", &providerConfig)
	return providerConfig
}

// GetName returns provider name
func (c *ThreePlayProvider) GetName() string {
	return providerName
}

// Download downloads captions file from specified type
func (c *ThreePlayProvider) Download(job *database.Job, captionsType string) ([]byte, error) {
	callParams := threeplay.CallParams{APIKey: c.config.APIKeyByJobType[job.JobType]}
	transcript, err := c.GetTranscriptText(job.GetProviderID(), "", types.CaptionsFormat(captionsType), callParams)
	if err != nil {
		return nil, err
	}
	return []byte(transcript), nil
}

// GetProviderJob returns a 3play file
func (c *ThreePlayProvider) GetProviderJob(job *database.Job) (*database.ProviderJob, error) {
	callParams := threeplay.CallParams{APIKey: c.config.APIKeyByJobType[job.JobType]}
	file, err := c.GetTranscriptInfo(job.GetProviderID(), callParams)
	if err != nil {
		return nil, err
	}

	providerJob := &database.ProviderJob{
		ID:          strconv.Itoa(file.ID),
		Status:      file.Status,
		Details:     file.Type,
		Cancellable: file.Cancellable,
	}
	return providerJob, nil
}

// DispatchJob sends a video file to 3play for transcription and captions generation or generates a expiring editing link
// when the media_file_url param is provided
//nolint:funlen
func (c *ThreePlayProvider) DispatchJob(job *database.Job) error {
	jobLogger := c.logger.WithFields(log.Fields{"JobID": job.ID, "Provider": job.Provider})
	callParams := threeplay.CallParams{APIKey: c.config.APIKeyByJobType[job.JobType]}

	// Review job route
	if transcriptID, ok := job.ProviderParams["transcript_id"]; ok {
		hoursInt := 2
		var err error
		if hoursUntilExpiration, ok := job.ProviderParams["hours_until_expiration"]; ok {
			hoursInt, err = strconv.Atoi(hoursUntilExpiration)
			if err != nil {
				jobLogger.WithError(err).Error("Could not convert hours until expiration")
			}
		}
		reviewURL, err := c.GetEditingLink(transcriptID, hoursInt, callParams)
		if err != nil {
			jobLogger.WithError(err).Error("Could not generate review url")
			return err
		}

		job.ProviderParams["ProviderID"] = transcriptID
		job.ProviderParams["ReviewURL"] = reviewURL
		return nil
	}

	// Creation job route
	query := url.Values{}
	turnaroundLevel := "asr"
	callbackURL := ""
	for k, v := range job.ProviderParams {
		switch k {
		case "turnaround_level_id":
			turnaroundLevel = v
		case "callback":
			url, err := url.Parse(v)
			if err != nil {
				break
			}
			query := url.Query()
			query.Add("job_id", job.ID)
			url.RawQuery = query.Encode()
			callbackURL = url.String()
		default:
			query.Add(k, v)
		}
	}
	fileID, err := c.UploadFileFromURL(query, callParams)

	if err != nil {
		jobLogger.Error("Failed to upload file to 3Play", err)
		return err
	}

	transcriptResponse, err := c.OrderTranscript(strconv.Itoa(fileID), callbackURL, turnaroundLevel, callParams)
	if err != nil {
		jobLogger.Error("Failed to order caption", err)
		return err
	}

	job.ProviderParams["ProviderID"] = strconv.Itoa(transcriptResponse.ID)
	return nil
}

// CancelJob cancels a job if it is in a cancellable state
func (c *ThreePlayProvider) CancelJob(job *database.Job) (bool, error) {
	callParams := threeplay.CallParams{APIKey: c.config.APIKeyByJobType[job.JobType]}
	providerJob, err := c.GetProviderJob(job)
	if err != nil {
		return false, err
	}
	if providerJob.Cancellable {
		err = c.CancelTranscript(providerJob.ID, callParams)
		if err != nil {
			return providerJob.Cancellable, err
		}
		return providerJob.Cancellable, nil
	}
	return providerJob.Cancellable, errors.New("job is not cancellable")
}
