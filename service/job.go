package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/NYTimes/gizmo/web"
	"github.com/NYTimes/video-captions-api/providers"
	log "github.com/Sirupsen/logrus"
)

// CaptionsError wraps error messages and uniforms json response
type CaptionsError struct {
	Message string `json:"error"`
}

// Error implements the error interface
func (e CaptionsError) Error() string {
	return e.Message
}

// GetJobs returns all the Jobs associated with a ParentID
func (s *CaptionsService) GetJobs(r *http.Request) (int, interface{}, error) {
	parentID := web.Vars(r)["id"]
	jobs, err := s.client.GetJobs(parentID)
	if err != nil {
		return http.StatusNotFound, nil, CaptionsError{err.Error()}
	}
	return http.StatusOK, jobs, nil
}

// GetJob returns a Job given its ID
func (s *CaptionsService) GetJob(r *http.Request) (int, interface{}, error) {
	id := web.Vars(r)["id"]
	// TODO: on the 3play client, we should look at the errors field and check for not_found errors at least
	job, err := s.client.GetJob(id)
	if err != nil {
		return http.StatusNotFound, nil, CaptionsError{err.Error()}
	}
	return http.StatusOK, job, nil
}

// CreateJob create a Job
func (s *CaptionsService) CreateJob(r *http.Request) (int, interface{}, error) {
	requestLogger := s.logger.WithFields(log.Fields{
		"Handler": "CreateJob",
		"Method":  r.Method,
		"URI":     r.RequestURI,
	})
	jobParams := providers.JobParams{}
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		requestLogger.WithError(err).Error("Could not read request body: ")
		return http.StatusBadRequest, nil, CaptionsError{err.Error()}
	}
	err = json.Unmarshal(data, &jobParams)

	if err != nil {
		requestLogger.WithError(err).Error("Could not create job from request body")
		return http.StatusBadRequest, nil, CaptionsError{"Malformed parameters"}
	}

	if jobParams.MediaURL == "" {
		requestLogger.WithError(err).Error("Tried to create a job without a media url")
		return http.StatusBadRequest, nil, CaptionsError{"Please provide a media_url"}
	}

	job := providers.NewJobFromParams(jobParams)
	err = s.client.DispatchJob(job)
	if err != nil {
		return http.StatusInternalServerError, nil, CaptionsError{err.Error()}
	}

	return http.StatusCreated, job, nil
}
