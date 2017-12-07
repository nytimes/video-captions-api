package service

import (
	"testing"

	"bytes"
	"encoding/json"
	"net/http"

	"fmt"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/stretchr/testify/assert"

	"io/ioutil"
	"net/http/httptest"
)

func TestCreateJob(t *testing.T) {
	assert := assert.New(t)
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/2019/video.mp4",
		Provider: "test-provider",
	}
	jobBytes, _ := json.Marshal(job)
	r, _ := http.NewRequest("POST", "/captions", bytes.NewReader(jobBytes))
	status, resultJob, err := service.CreateJob(r)
	job = resultJob.(*database.Job)
	assert.Nil(err)
	assert.Equal(201, status)
	expectedFileName := fmt.Sprintf("video_%s.vtt", job.ID)
	assert.Equal(job.Outputs, []database.JobOutput{
		{Filename: expectedFileName, Type: "vtt"},
	})
}

func TestCreateJobQueryString(t *testing.T) {
	assert := assert.New(t)
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/2019/video.mp4?v=19847248429429&a=abc123&GoogleAccessId=user%40exame.iam.gserviceaccount.com",
		Provider: "test-provider",
	}
	jobBytes, _ := json.Marshal(job)
	r, _ := http.NewRequest("POST", "/captions", bytes.NewReader(jobBytes))
	status, resultJob, err := service.CreateJob(r)
	job = resultJob.(*database.Job)
	assert.Nil(err)
	assert.Equal(201, status)
	expectedFileName := fmt.Sprintf("video_%s.vtt", job.ID)
	assert.Equal(job.Outputs, []database.JobOutput{
		{Filename: expectedFileName, Type: "vtt"},
	})
}

func TestCreateUploadJob(t *testing.T) {
	assert := assert.New(t)
	service, client := createCaptionsService()
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ID:          "123",
		CaptionFile: database.UploadedFile{[]byte("captions"), "caption.net"},
		Provider:    "test-provider",
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
	assert.Equal("Please provide a media_url or caption_file", err.Error())
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
		Done:     false,
		Status:   "in_progress",
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
	r2, _ := http.NewRequest("POST", urlStr, nil)
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
	r, _ := http.NewRequest("POST", "/jobs/404/cancel", nil)
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
		Done:     true,
	}
	client.DB.StoreJob(job)
	server.Register(service)
	r, _ := http.NewRequest("POST", "/jobs/123/cancel", nil)
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

func TestDownload(t *testing.T) {
	service, client := createCaptionsService()
	server := server.NewSimpleServer(&server.Config{})
	assert := assert.New(t)
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
	}
	client.DB.StoreJob(job)
	server.Register(service)
	r, _ := http.NewRequest("GET", "/jobs/123/download/vtt", bytes.NewReader(nil))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	assert.Equal(200, w.Code)
	body, _ := ioutil.ReadAll(w.Body)
	assert.Equal("WEBVTT\n\nNOTE Paragraph\n\n00:00:09.240 --> 00:00:11.010\nWe're all talking\nabout the Iowa caucuses", string(body))
}

func TestDownloadMissingCaption(t *testing.T) {
	service, client := createCaptionsService()
	server := server.NewSimpleServer(&server.Config{})
	assert := assert.New(t)
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
	}
	client.DB.StoreJob(job)
	server.Register(service)
	r, _ := http.NewRequest("GET", "/jobs/456/download/srt", bytes.NewReader(nil))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	assert.Equal(404, w.Code)
	body, _ := ioutil.ReadAll(w.Body)
	assert.Equal("", string(body))
}

func TestDownloadBadRequest(t *testing.T) {
	service, client := createCaptionsService()
	server := server.NewSimpleServer(&server.Config{})
	assert := assert.New(t)
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
	}
	client.DB.StoreJob(job)
	server.Register(service)
	r, _ := http.NewRequest("GET", "/jobs/123/download", bytes.NewReader(nil))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	assert.Equal(404, w.Code)
	body, _ := ioutil.ReadAll(w.Body)
	assert.Equal("404 page not found\n", string(body))
}

func TestTranscript(t *testing.T) {
	service, client := createCaptionsService()
	server := server.NewSimpleServer(&server.Config{})
	assert := assert.New(t)
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
	}
	client.DB.StoreJob(job)
	server.Register(service)
	r, _ := http.NewRequest("GET", "/jobs/123/transcript/vtt", bytes.NewReader(nil))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	assert.Equal(200, w.Code)
	body, _ := ioutil.ReadAll(w.Body)
	assert.Equal("We're all talking about the Iowa caucuses", string(body))
}

func TestTranscriptMissingCaption(t *testing.T) {
	service, client := createCaptionsService()
	server := server.NewSimpleServer(&server.Config{})
	assert := assert.New(t)
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
	}
	client.DB.StoreJob(job)
	server.Register(service)
	r, _ := http.NewRequest("GET", "/jobs/456/transcript/vtt", bytes.NewReader(nil))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	assert.Equal(404, w.Code)
	body, _ := ioutil.ReadAll(w.Body)
	assert.Equal("", string(body))
}

func TestTranscriptBadRequest(t *testing.T) {
	service, client := createCaptionsService()
	server := server.NewSimpleServer(&server.Config{})
	assert := assert.New(t)
	service.AddProvider(fakeProvider{logger: client.Logger})
	job := &database.Job{
		ID:       "123",
		MediaURL: "http://vp.nyt.com/video.mp4",
		Provider: "test-provider",
	}
	client.DB.StoreJob(job)
	server.Register(service)
	r, _ := http.NewRequest("GET", "/jobs/123/transcript/wrong", bytes.NewReader(nil))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	assert.Equal(400, w.Code)
}
