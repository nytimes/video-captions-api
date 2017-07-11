package service

import (
	"testing"

	"bytes"
	"encoding/json"
	"net/http"

	"github.com/NYTimes/video-captions-api/providers"

	"github.com/stretchr/testify/assert"
)

func TestCreateJob(t *testing.T) {
	assert := assert.New(t)
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &providers.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
	}
	jobBytes, _ := json.Marshal(job)
	r, _ := http.NewRequest("POST", "/captions", bytes.NewReader(jobBytes))
	status, resultJob, err := service.CreateJob(r)
	job = resultJob.(*providers.Job)
	assert.Nil(err)
	assert.Equal(201, status)
	assert.Equal(job.ParentID, "123")
}

func TestCreateJobNoMediaURL(t *testing.T) {
	assert := assert.New(t)
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &providers.Job{
		ID:       "123",
		Provider: "test-provider",
	}
	jobBytes, _ := json.Marshal(job)
	r, _ := http.NewRequest("POST", "/captions", bytes.NewReader(jobBytes))
	status, resultJob, err := service.CreateJob(r)
	assert.NotNil(err)
	assert.Equal("Please provide a media_url", err.Error())
	assert.Nil(resultJob)
	assert.Equal(400, status)
}

func TestCreateJobDispatchError(t *testing.T) {
	assert := assert.New(t)
	service, client := createCaptionsService()
	service.AddProvider(brokenProvider{logger: client.Logger})
	job := &providers.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "broken-provider",
	}
	jobBytes, _ := json.Marshal(job)
	r, _ := http.NewRequest("POST", "/captions", bytes.NewReader(jobBytes))
	status, _, err := service.CreateJob(r)
	assert.NotNil(err)
	assert.Equal("Error dispatching Job", err.Error())
	assert.Equal(500, status)
}

func TestCreateJobInvalidBody(t *testing.T) {
	assert := assert.New(t)
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{logger: client.Logger})
	r, _ := http.NewRequest("POST", "/captions", bytes.NewReader([]byte("not json")))
	status, resultJob, err := service.CreateJob(r)
	assert.NotNil(err)
	assert.Equal("Malformed parameters", err.Error())
	assert.Nil(resultJob)
	assert.Equal(400, status)
}

func TestGetJob404(t *testing.T) {
	assert := assert.New(t)
	service, _ := createCaptionsService()
	r, _ := http.NewRequest("GET", "/jobs/404", nil)
	status, _, err := service.GetJob(r)
	assert.Equal(status, 404)
	assert.Equal("Job doesn't exist", err.Error())
}
