package service

import (
	"testing"

	"bytes"
	"encoding/json"
	"net/http"

	"github.com/NYTimes/video-captions-api/database"
	"github.com/stretchr/testify/assert"
	"fmt"
	"github.com/NYTimes/gizmo/server"

	"net/http/httptest"
)

func TestCreateJob(t *testing.T) {
	assert := assert.New(t)
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
	}
	jobBytes, _ := json.Marshal(job)
	r, _ := http.NewRequest("POST", "/captions", bytes.NewReader(jobBytes))
	status, resultJob, err := service.CreateJob(r)
	job = resultJob.(*database.Job)
	assert.Nil(err)
	assert.Equal(201, status)
}

func TestCreateJobNoMediaURL(t *testing.T) {
	assert := assert.New(t)
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
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
	job := &database.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "broken-provider",
	}
	jobBytes, _ := json.Marshal(job)
	r, _ := http.NewRequest("POST", "/captions", bytes.NewReader(jobBytes))
	status, _, err := service.CreateJob(r)
	assert.NotNil(err)
	assert.Equal("Error dispatching Job: provider error", err.Error())
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
	server := server.NewSimpleServer(&server.Config{})
	service, _ := createCaptionsService()
	server.Register(service)
	r, _ := http.NewRequest("GET", "/jobs/404", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	var jobBody map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&jobBody)
	if err != nil {
		t.Errorf("%s: unable to JSON decode response body: %s", w.Body, err)
	}
	assert.Equal(w.Code, 404)
	assert.Equal("Job doesn't exist", jobBody["error"])
}

func TestCancelJob(t *testing.T) {
	assert := assert.New(t)
	server := server.NewSimpleServer(&server.Config{})
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ParentID: "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
		Done: false,
		Status: "in_progress",
	}
	server.Register(service)

	jobBytes, _ := json.Marshal(job)
	r, _ := http.NewRequest("POST", "/captions", bytes.NewReader(jobBytes))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	assert.Equal(201, w.Code)
	var captionsBody map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&captionsBody)
	if err != nil {
		t.Errorf("%s: unable to JSON decode response body: %s", w.Body, err)
	}

	urlStr := fmt.Sprintf("/jobs/%v/cancel", captionsBody["id"])
	r2, _ := http.NewRequest("GET", urlStr, nil)
	w2 := httptest.NewRecorder()
	server.ServeHTTP(w2, r2)
	assert.Equal(200, w2.Code)
	var cancelBody map[string]interface{}
	err = json.NewDecoder(w2.Body).Decode(&cancelBody)
	if err != nil {
		t.Errorf("%s: unable to JSON decode response body: %s", w2.Body, err)
	}
	assert.Nil(cancelBody)
}

func TestCancelJob404(t *testing.T) {
	assert := assert.New(t)
	server := server.NewSimpleServer(&server.Config{})
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{logger: client.Logger})
	server.Register(service)
	r, _ := http.NewRequest("GET", "/jobs/404/cancel", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)

	var cancelBody map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&cancelBody)
	if err != nil {
		t.Errorf("%s: unable to JSON decode response body: %s", w.Body, err)
	}

	assert.Equal(w.Code, 404)
	assert.Equal("Job doesn't exist", cancelBody["error"])
}

func TestCancelJobDone(t *testing.T) {
	service, client := createCaptionsService()
	server := server.NewSimpleServer(&server.Config{})
	assert := assert.New(t)
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
		Done: true,
	}
	client.DB.StoreJob(job)
	server.Register(service)
	r, _ := http.NewRequest("GET", "/jobs/123/cancel", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	assert.Equal(409, w.Code)
	var cancelBody map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&cancelBody)
	if err != nil {
		t.Errorf("%s: unable to JSON decode response body: %s", w.Body, err)
	}
	assert.Equal("Cannot cancel a job that is already done", cancelBody["error"])
}
