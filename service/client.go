package service

import (
	"errors"
	"fmt"
	"sort"

	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	log "github.com/Sirupsen/logrus"
)

// Client CaptionsService client
type Client struct {
	Providers map[string]providers.Provider
	DB        database.DB
	Logger    *log.Logger
	Storage   Storage
}

// GetJobs gets all jobs associated with a ParentID
func (c Client) GetJobs(parentID string) ([]*database.JobSummary, error) {
	jobs, err := c.DB.GetJobs(parentID)
	if err != nil {
		c.Logger.Errorf("Error loading jobs from DB for parent ID: %s", parentID)
		return nil, err
	}
	sort.Sort(database.ByCreatedAt(jobs))
	summaries := make([]*database.JobSummary, len(jobs))
	for i, job := range jobs {
		summaries[i] = &database.JobSummary{ID: job.ID, CreatedAt: job.CreatedAt}
	}
	return summaries, nil
}

// GetJob gets a job by ID
func (c Client) GetJob(jobID string) (*database.Job, error) {
	job, err := c.DB.GetJob(jobID)
	if err != nil {
		c.Logger.Error("Could not find Job in database")
		return nil, err
	}

	if job.Done {
		return job, nil
	}

	providerID := job.ProviderParams["ProviderID"]
	fields := log.Fields{"JobID": jobID, "Provider": job.Provider, "ProviderID": providerID}
	jobLogger := c.Logger.WithFields(fields)
	provider := c.Providers[job.Provider]
	jobLogger.Info("Fetching job from Provider")
	providerJob, err := provider.GetProviderJob(providerID)
	if err != nil {
		jobLogger.Error("error getting job from provider", err)
		return nil, err
	}

	params := providerJob.Params

	shouldUpdate := false

	for k, v := range params {
		if params[k] != job.ProviderParams[k] {
			job.ProviderParams[k] = v
			shouldUpdate = true
		}
	}

	if job.UpdateStatus(providerJob.Status, providerJob.Details) || shouldUpdate {
		err = c.DB.UpdateJob(jobID, job)
	}

	if job.Status == "delivered" && !job.Done {
		jobLogger.Info("Job is ready on the provider, downloading")
		for i, output := range job.Outputs {
			data, err := provider.Download(providerID, output.Type)
			if err != nil {
				jobLogger.WithError(err).Error("Failed to download file")
				return job, nil
			}
			jobLogger.Info("Download done, storing")
			dest, err := c.Storage.Store(data, fmt.Sprintf("%s/%s", job.Provider, output.Filename))
			if err != nil {
				jobLogger.WithError(err).Error("Failed to store file")
				return job, nil
			}
			job.Outputs[i].URL = dest
		}
		job.Done = true
		err = c.DB.UpdateJob(jobID, job)
	}
	return job, err
}

// DispatchJob dispatches a Job given an existing Provider
func (c Client) DispatchJob(job *database.Job) error {
	provider := c.Providers[job.Provider]
	jobLogger := c.Logger.WithFields(log.Fields{"JobID": job.ID, "Provider": job.Provider})
	if provider == nil {
		jobLogger.Error("Provider not found")
		return errors.New("Provider not found")
	}

	jobLogger.Info("Dispatching job to provider")
	err := provider.DispatchJob(job)
	if err != nil {
		jobLogger.Errorf("Error dispatching job to provider: %v", err)
		return fmt.Errorf("Error dispatching Job: %v", err)
	}
	jobLogger.Info("Storing job in DB")
	_, err = c.DB.StoreJob(job)
	if err != nil {
		jobLogger.Errorf("Error storing job in DB: %v", err)
		return fmt.Errorf("Error storing Job: %v", err)
	}
	return nil
}

// CancelJob cancels a job by ID
func (c Client) CancelJob(jobID string) (bool, error) {
	job, err := c.DB.GetJob(jobID)
	if err != nil {
		c.Logger.Error("Could not find Job in database")
		return false, err
	}

	if job.Done {
		c.Logger.Error("Cannot cancel a job that is already done")
		return false, nil
	}

	job.Status = "canceled"
	job.Done = true

	err = c.DB.UpdateJob(jobID, job)

	return true, err
}
