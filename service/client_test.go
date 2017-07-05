package service

import (
	"testing"

	"github.com/NYTimes/video-captions-api/providers"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"reflect"
)

func TestGetJob(t *testing.T) {
	service, client := createCaptionsService()
	assert := assert.New(t)
	service.AddProvider(fakeProvider{logger: log.New()})
	job := providers.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
	}
	client.DB.StoreJob(job)
	resultJob, _ := client.GetJob(job.ID)
	assert.Equal(job.ID, resultJob.ID)
	assert.Equal(job.MediaURL, resultJob.MediaURL)
}

func TestDispatchJobNoProvider(t *testing.T) {
	_, client := createCaptionsService()
	_, err := client.DispatchJob(providers.Job{Provider: "wrong-provider"})
	assert.Equal(t, "Provider not found", err.Error())
}

func TestGetJobs(t *testing.T) {
	parentID := "Mom"

	service, client := createCaptionsService()

	service.AddProvider(fakeProvider{logger: log.New()})
	job1 := providers.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video1.mp4",
		Provider: "test-provider1",
		ParentID: parentID,
	}
	client.DB.StoreJob(job1)
	job2 := providers.Job{
		ID:       "456",
		MediaURL: "http://vp.nyt.com/video2.mp4",
		Provider: "test-provider2",
		ParentID: parentID,
	}
	client.DB.StoreJob(job2)

	resultJob, _ := client.GetJobs(parentID)

	assert := assert.New(t)
	assert.NotNil(resultJob)
	assert.Len(resultJob, 2)
	if !reflect.DeepEqual(job1, resultJob[0]) {
		t.Errorf("The first job did not match\nExpected: %#v\nGot:  %#v", job1, resultJob[0])
	}
	if !reflect.DeepEqual(job2, resultJob[1]) {
		t.Errorf("The second job did not match\nExpected: %#v\nGot:  %#v", job2, resultJob[1])
	}
}

func TestProviderJobError(t *testing.T) {
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{
		logger: log.New(),
		params: map[string]bool{
			"jobError": true,
			"jobStatus": false,
		},
	})
	job := providers.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
	}
	client.DB.StoreJob(job)
	_, err := client.GetJob("123")

	assert := assert.New(t)
	assert.NotNil(err)
	assert.EqualValues(err.Error(), "oh no")

}

func TestProviderStatusError(t *testing.T) {
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{
		logger: log.New(),
		params: map[string]bool{
			"jobError": false,
			"jobStatus": true,
		},
	})
	job := providers.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
	}
	client.DB.StoreJob(job)
	resultJob, _ := client.GetJob("123")
	assert := assert.New(t)
	assert.NotNil(resultJob.Status)
	assert.EqualValues(resultJob.Status, "My status")
}