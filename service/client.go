package service

import (
	"errors"
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
func (c Client) GetJobs(parentID string) ([]database.Job, error) {
	jobs, err := c.DB.GetJobs(parentID)
	if err != nil {
		c.Logger.Error("Error loading jobs from DB", parentID)
		return nil, err
	}
	sort.Sort(database.ByCreatedAt(jobs))
	return jobs, nil
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
	providerJob, err := provider.GetJob(providerID)
	if err != nil {
		jobLogger.Error("error getting job from provider", err)
		return nil, err
	}

	if job.UpdateStatus(providerJob.Status) {
		err = c.DB.UpdateJob(jobID, job)
	}

	if job.Status == "delivered" && !job.Done {
		jobLogger.Info("Job is ready on the provider, downloading")
		// TODO: do this async so we dont block here, once the captions are ready,
		// we start the download/upload update the status to something like
		// "storing"/"downloading" and return the response to the user.
		// once the goroutines are done downloading/storing, mark the job as
		// done. maybe spawn one goroutine per output?
		for i, output := range job.Outputs {
			data, err := provider.Download(providerID, output.Type)
			if err != nil {
				jobLogger.WithError(err).Error("Failed to download file")
				return job, nil
			}
			jobLogger.Info("Download done, storing")
			dest, err := c.Storage.Store(data, output.Filename)
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
		jobLogger.Error("Error dispatching job to provider", err)
		return errors.New("Error dispatching Job")
	}
	jobLogger.Info("Storing job in DB")
	_, err = c.DB.StoreJob(job)
	if err != nil {
		jobLogger.Error("Error storing job in DB", err)
		return errors.New("Error storing Job")
	}
	return nil
}
