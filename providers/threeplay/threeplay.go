package threeplay

import (
	"net/url"
	"strconv"

	"github.com/NYTimes/threeplay"
	"github.com/NYTimes/video-captions-api/providers"
)

const providerName string = "3play"

type Provider struct {
	*threeplay.Client
}

func New(APIKey, APISecret string) providers.Provider {
	return &Provider{
		threeplay.NewClient(APIKey, APISecret),
	}
}

func (c *Provider) GetJob(id string) (*providers.Job, error) {
	i, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	file, err := c.GetFile(uint(i))
	if err != nil {
		return nil, err
	}

	job := &providers.Job{
		ID:       strconv.FormatUint(uint64(file.ID), 10),
		Status:   file.State,
		Provider: providerName,
	}
	return job, nil
}

func (c *Provider) DispatchJob(job *providers.Job) (*providers.Job, error) {
	// TODO: we need to parse job.ProviderParams to query
	query, _ := url.ParseQuery("for_asr=1")

	fileID, err := c.UploadFileFromURL(job.MediaURL, query)

	if err != nil {
		return nil, err
	}
	job.Provider = "3play"
	job.ID = fileID
	job.Status = "processing"

	return job, nil
}
