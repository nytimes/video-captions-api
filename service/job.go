package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
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
	Language       string                  `json:"language"`
}

func newJobFromParams(newJob jobParams) (*database.Job, error) {
	outputs := make([]database.JobOutput, 0)
	mediaFile := filepath.Base(newJob.MediaURL)
	name := strings.TrimSuffix(mediaFile, filepath.Ext(mediaFile))
	id, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("could not create job id: %v", err)
	}

	for _, outputType := range newJob.OutputTypes {
		fileName := fmt.Sprintf("%s_%s.%s", name, id.String(), outputType)
		outputs = append(outputs, database.JobOutput{Type: outputType, Filename: fileName})
	}

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
		Language:       newJob.Language,
	}, nil
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

// CancelJob cancels a given Job by its ID
func (s *CaptionsService) CancelJob(r *http.Request) (int, interface{}, error) {
	id := web.Vars(r)["id"]
	canceled, err := s.client.CancelJob(id)
	if err != nil {
		return http.StatusNotFound, nil, captionsError{err.Error()}
	}
	if !canceled {
		return http.StatusConflict, nil, captionsError{"Cannot cancel a job that is already done"}
	}
	return http.StatusOK, nil, nil
}

// CreateJob create a Job
func (s *CaptionsService) CreateJob(r *http.Request) (int, interface{}, error) {
	requestLogger := s.logger.WithFields(log.Fields{
		"Handler": "CreateJob",
		"Method":  r.Method,
		"URI":     r.RequestURI,
	})
	params := jobParams{
		Language:       "en",
		OutputTypes:    []string{"vtt"},
		ProviderParams: make(database.ProviderParams),
	}
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

	job, err := newJobFromParams(params)
	if err != nil {
		requestLogger.WithError(err).Error("could not create job from parameters")
		return http.StatusInternalServerError, nil, captionsError{err.Error()}
	}

	err = s.client.DispatchJob(job)
	if err != nil {
		requestLogger.WithError(err).Error("could not dispatch job")
		return http.StatusInternalServerError, nil, captionsError{err.Error()}
	}

	return http.StatusCreated, job, nil
}

// DownloadCaption downloads a caption in the specified format
func (s *CaptionsService) DownloadCaption(w http.ResponseWriter, r *http.Request) {
	id := web.Vars(r)["id"]
	captionFormat := web.Vars(r)["captionFormat"]

	defer r.Body.Close()

	captionFile, err := s.client.DownloadCaption(id, captionFormat)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Set("Content-Type", fmt.Sprintf("text/%s; charset=utf-8", captionFormat))
	w.WriteHeader(http.StatusOK)
	w.Write(captionFile)
}

// DownloadCaption downloads a caption in the specified format
func (s *CaptionsService) GetTranscript(w http.ResponseWriter, r *http.Request) {
	id := web.Vars(r)["id"]
	captionFormat := web.Vars(r)["captionFormat"]

	defer r.Body.Close()

	captionFile, err := s.client.DownloadCaption(id, captionFormat)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	transcript, err := s.client.GenerateTranscript(captionFile, captionFormat)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(transcript))
}
