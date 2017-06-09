package providers

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/NYTimes/threeplay"
)

const providerName string = "3play"

type ThreePlayProvider struct {
	*threeplay.Client
}

func NewThreePlay(APIKey, APISecret string) Provider {
	return &ThreePlayProvider{
		threeplay.NewClient(APIKey, APISecret),
	}
}

func (c *ThreePlayProvider) GetName() string {
	return providerName
}

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

func (c *ThreePlayProvider) DispatchJob(job Job) (Job, error) {
	fmt.Println("Dispatching job to 3Play", job.ID)
	// TODO: we need to parse job.ProviderParams to query
	query, _ := url.ParseQuery("for_asr=1")

	fileID, err := c.UploadFileFromURL(job.MediaURL, query)

	if err != nil {
		return Job{}, err
	}

	job.ProviderID = fileID

	return job, nil
}
