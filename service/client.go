package service

import (
	"errors"

	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	log "github.com/Sirupsen/logrus"
)

// Client CaptionsService client
type Client struct {
	Providers map[string]providers.Provider
	DB        database.DB
	Logger    *log.Logger
}

// GetJobs gets all jobs associated with a ParentID
func (c Client) GetJobs(parentID string) ([]providers.Job, error) {
	jobs, err := c.DB.GetJobs(parentID)
	if err != nil {
		c.Logger.Error("Error loading jobs from DB", parentID)
		return nil, err
	}
	return jobs, nil
}

// GetJob gets a job by ID
func (c Client) GetJob(jobID string) (providers.Job, error) {
	job, err := c.DB.GetJob(jobID)
	if err != nil {
		c.Logger.Error("Could not find Job in database")
		return job, err
	}

	jobLogger := c.Logger.WithFields(log.Fields{"JobID": jobID, "Provider": job.Provider})
	provider := c.Providers[job.Provider]
	jobLogger.Info("Fetching job from Provider")
	providerJob, err := provider.GetJob(job.ProviderID)
	if err != nil {
		jobLogger.Error("error getting job from provider", err)
		return job, err
	}

	if providerJob.Status != job.Status {
		jobLogger.Info("Updating job status: ", job.Status, "->", providerJob.Status)
		job.Status = providerJob.Status
		err = c.DB.UpdateJob(jobID, job)
	} else {
		jobLogger.Info("No job status update")
	}

	return job, err
}

// DispatchJob dispatches a Job given an existing Provider
func (c Client) DispatchJob(job providers.Job) (providers.Job, error) {
	provider := c.Providers[job.Provider]
	jobLogger := c.Logger.WithFields(log.Fields{"JobID": job.ID, "Provider": job.Provider})
	if provider == nil {
		jobLogger.Error("Provider not found")
		return providers.Job{}, errors.New("Provider not found")
	}
	job.Status = "processing"

	jobLogger.Info("Dispatching job to provider")
	job, err := provider.DispatchJob(job)
	if err != nil {
		jobLogger.Error("Error dispatching job to provider", err)
		return providers.Job{}, errors.New("Error dispatching Job")
	}
	jobLogger.Info("Storing job in DB")
	_, err = c.DB.StoreJob(job)
	if err != nil {
		jobLogger.Error("Error storing job in DB", err)
		return providers.Job{}, errors.New("Error storing Job")
	}
	return job, err
}
