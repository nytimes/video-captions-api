package providers

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/NYTimes/threeplay"
)

const providerName string = "3play"

// ThreePlayProvider is a 3play client that implements the Provider interface
type ThreePlayProvider struct {
	*threeplay.Client
}

// NewThreePlay creates a ThreePlayProvider instance
func NewThreePlay(APIKey, APISecret string) Provider {
	return &ThreePlayProvider{
		threeplay.NewClient(APIKey, APISecret),
	}
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
