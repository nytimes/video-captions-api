package service

import (
	"errors"
	"fmt"

	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
)

// Client CaptionsService client
type Client struct {
	Providers map[string]providers.Provider
	DB        database.DB
}

// GetJob gets a job by ID
func (c Client) GetJob(id string) (providers.Job, error) {
	job, err := c.DB.GetJob(id)
	if err != nil {
		return job, err
	}

	fmt.Println("Fetching job status")
	provider := c.Providers[job.Provider]
	providerJob, err := provider.GetJob(job.ProviderID)
	if err != nil {
		return job, err
	}

	fmt.Println("updating job status:", job.Status, "->", providerJob.Status)
	job.Status = providerJob.Status

	fmt.Println("updating job in DB")
	err = c.DB.UpdateJob(id, job)
	return job, err
}

// DispatchJob dispatches a Job given an existing Provider
func (c Client) DispatchJob(job providers.Job) (providers.Job, error) {
	provider := c.Providers[job.Provider]
	if provider == nil {
		return providers.Job{}, errors.New("Provider not found")
	}
	job.Status = "processing"

	fmt.Println("Got provider, storing job in DB")
	job, err := provider.DispatchJob(job)
	if err != nil {
		fmt.Println("Service Client Error:", err)
		return providers.Job{}, errors.New("Error dispatching Job")
	}
	fmt.Println("Job dispatched to provider", provider)
	_, err = c.DB.StoreJob(job)
	if err != nil {
		fmt.Println("Service Client Error:", err)
		return providers.Job{}, errors.New("Error storing Job")
	}
	return job, err
}
