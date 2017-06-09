package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/NYTimes/gizmo/web"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/satori/go.uuid"
)

func (s *SimpleService) GetJob(r *http.Request) (int, interface{}, error) {
	id := web.Vars(r)["id"]
	fmt.Println("GetJob", id)
	// TODO: on the 3play client, we should look at the errors field and check for not_found errors at least
	file, err := s.client.GetJob(id)
	if err != nil {
		return http.StatusNotFound, nil, err
	}
	return http.StatusOK, file, nil
}

func (s *SimpleService) CreateJob(r *http.Request) (int, interface{}, error) {
	fmt.Println("CreateJob")
	var job providers.Job
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}
	err = json.Unmarshal(data, &job)

	if err != nil {
		return http.StatusBadRequest, nil, errors.New("Malformed parameters")
	}

	mediaURL := job.MediaURL

	if mediaURL == "" {
		return http.StatusBadRequest, nil, errors.New("Please provide a media_url")
	}

	job.ID = uuid.NewV4().String()

	job, err = s.client.DispatchJob(job)
	if err != nil {
		fmt.Println("Error", err)
		return http.StatusInternalServerError, err, err
	}

	return http.StatusCreated, job, nil
}
