package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/NYTimes/gizmo/web"
	"github.com/NYTimes/video-captions-api/database"
	log "github.com/Sirupsen/logrus"
	uuid "github.com/nu7hatch/gouuid"
)

type captionsError struct {
	Message string `json:"error"`
}

type jobParams struct {
	ParentID       string                  `json:"parent_id"`
	MediaURL       string                  `json:"media_url"`
	Provider       string                  `json:"provider"`
	ProviderParams database.ProviderParams `json:"provider_params"`
	OutputTypes    []string                `json:"output_types"`
}

func NewJobFromParams(newJob jobParams) *database.Job {
	outputs := make([]database.JobOutput, 0)
	for _, outputType := range newJob.OutputTypes {
		outputs = append(outputs, database.JobOutput{Type: outputType})
	}
	id, _ := uuid.NewV4()
	return &database.Job{
		ID:       id.String(),
		ParentID: newJob.ParentID,
		//TODO: put all possible status under a type/consts so we dont use strings everywhere
		Status:         "processing",
		MediaURL:       newJob.MediaURL,
		Provider:       newJob.Provider,
		ProviderParams: newJob.ProviderParams,
		CreatedAt:      time.Now(),
		Outputs:        outputs,
		Done:           false,
	}
}

// Error implements the error interface
func (e captionsError) Error() string {
	return e.Message
}

// GetJobs returns all the Jobs associated with a ParentID
func (s *CaptionsService) GetJobs(r *http.Request) (int, interface{}, error) {
	parentID := web.Vars(r)["id"]
	jobs, err := s.client.GetJobs(parentID)
	if err != nil {
		return http.StatusNotFound, nil, captionsError{err.Error()}
	}
	return http.StatusOK, jobs, nil
}

// GetJob returns a Job given its ID
func (s *CaptionsService) GetJob(r *http.Request) (int, interface{}, error) {
	id := web.Vars(r)["id"]
	// TODO: on the 3play client, we should look at the errors field and check for not_found errors at least
	job, err := s.client.GetJob(id)
	if err != nil {
		return http.StatusNotFound, nil, captionsError{err.Error()}
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
	params := jobParams{}
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		requestLogger.WithError(err).Error("Could not read request body: ")
		return http.StatusBadRequest, nil, captionsError{err.Error()}
	}
	err = json.Unmarshal(data, &params)

	if err != nil {
		requestLogger.WithError(err).Error("Could not create job from request body")
		return http.StatusBadRequest, nil, captionsError{"Malformed parameters"}
	}

	if params.MediaURL == "" {
		requestLogger.WithError(err).Error("Tried to create a job without a media url")
		return http.StatusBadRequest, nil, captionsError{"Please provide a media_url"}
	}

	job := NewJobFromParams(params)
	err = s.client.DispatchJob(job)
	if err != nil {
		return http.StatusInternalServerError, nil, captionsError{err.Error()}
	}

	return http.StatusCreated, job, nil
}
