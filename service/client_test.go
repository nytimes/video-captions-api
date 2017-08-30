package service

import (
	"testing"
	"time"

	"github.com/NYTimes/video-captions-api/database"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetJob(t *testing.T) {
	service, client := createCaptionsService()
	assert := assert.New(t)
	service.AddProvider(fakeProvider{logger: log.New()})
	job := &database.Job{
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
	err := client.DispatchJob(&database.Job{Provider: "wrong-provider"})
	assert.Equal(t, "Provider not found", err.Error())
}

func TestGetJobReady(t *testing.T) {
	assert := assert.New(t)
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{
		logger: log.New(),
		params: map[string]bool{
			"jobError":  false,
			"jobStatus": false,
			"jobDone":   true,
		},
	})
	job, _ := NewJobFromParams(jobParams{
		MediaURL:    "http://vp.nyt.com/video.mp4",
		ParentID:    "123",
		Provider:    "test-provider",
		OutputTypes: []string{"vtt", "srt"},
	})
	job.Status = "delivered"
	client.DB.StoreJob(job)

	resultJob, _ := client.GetJob(job.ID)
	assert.True(resultJob.Done)
	assert.Equal("somepath/test-provider/"+resultJob.Outputs[0].Filename, resultJob.Outputs[0].URL)
	assert.Equal("somepath/test-provider/"+resultJob.Outputs[1].Filename, resultJob.Outputs[1].URL)
}

func TestGetJobs(t *testing.T) {
	parentID := "Mom"

	service, client := createCaptionsService()

	service.AddProvider(fakeProvider{logger: log.New()})
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)
	job1 := &database.Job{
		ID:        "123",
		MediaURL:  "http://vp.nyt.com/video1.mp4",
		Provider:  "test-provider1",
		ParentID:  parentID,
		CreatedAt: today,
	}
	client.DB.StoreJob(job1)
	job2 := &database.Job{
		ID:        "456",
		MediaURL:  "http://vp.nyt.com/video2.mp4",
		Provider:  "test-provider2",
		ParentID:  parentID,
		CreatedAt: yesterday,
	}
	client.DB.StoreJob(job2)

	summaries, _ := client.GetJobs(parentID)

	assert := assert.New(t)
	assert.NotNil(summaries)
	assert.Len(summaries, 2)

	assert.Equal(job1.ID, summaries[1].ID)
	assert.Equal(job2.ID, summaries[0].ID)
}

func TestProviderJobError(t *testing.T) {
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{
		logger: log.New(),
		params: map[string]bool{
			"jobError":  true,
			"jobStatus": false,
		},
	})
	job := &database.Job{
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
			"jobError":  false,
			"jobStatus": true,
		},
	})
	job := &database.Job{
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
