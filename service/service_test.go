package service

import (
	"errors"
	"fmt"
	"testing"

	"reflect"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/config"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type fakeProvider struct {
	logger *log.Logger
	params map[string]bool
}

func (p fakeProvider) DispatchJob(job *database.Job) error {
	p.logger.Info("dispatching job")
	return nil
}

func (p fakeProvider) Download(_ string, _ string) ([]byte, error) {
	return []byte("captions"), nil
}

func (p fakeProvider) GetJob(id string) (*database.Job, error) {
	if p.params["jobError"] {
		return nil, errors.New("oh no")
	}
	if p.params["jobStatus"] {
		return &database.Job{Status: "My status"}, nil
	}
	if p.params["jobDone"] {
		job := NewJobFromParams(jobParams{
			MediaURL:    "http://vp.nyt.com/video.mp4",
			ParentID:    "123",
			Provider:    "test-provider",
			OutputTypes: []string{"vtt", "srt"},
		})
		job.Status = "delivered"
		return job, nil
	}
	p.logger.Info("fetching job", id)
	return &database.Job{}, nil
}

func (p fakeProvider) GetName() string {
	return "test-provider"
}

type brokenProvider fakeProvider

func (p brokenProvider) GetName() string {
	return "broken-provider"
}

func (p brokenProvider) Download(_ string, _ string) ([]byte, error) {
	return nil, errors.New("download error")
}

func (p brokenProvider) DispatchJob(job *database.Job) error {
	return errors.New("provider error")
}

func (p brokenProvider) GetJob(id string) (*database.Job, error) {
	p.logger.Info("fetching job", id)
	return nil, errors.New("failed to get job")
}

func createCaptionsService() (*CaptionsService, Client) {
	client := Client{
		Providers: make(map[string]providers.Provider),
		DB:        database.NewMemoryDatabase(),
		Logger:    log.New(),
		Storage:   ignoreStorage{},
	}
	service := &CaptionsService{
		client: client,
		logger: log.New(),
	}
	return service, client
}

type ignoreStorage struct{}

func (i ignoreStorage) Store(_ []byte, filename string) (string, error) {
	return fmt.Sprintf("somepath/%s", filename), nil
}

func TestAddProvider(t *testing.T) {
	assert := assert.New(t)
	service, client := createCaptionsService()

	service.AddProvider(fakeProvider{})
	provider := client.Providers["test-provider"]
	assert.NotNil(provider)
	assert.Equal(provider.GetName(), "test-provider")
}

func TestNewCaptionsService(t *testing.T) {
	logger := log.New()
	projectID := "My amazing captions project"
	providers := make(map[string]providers.Provider)
	cfg := config.CaptionsServiceConfig{
		Server:    &server.Config{},
		Logger:    logger,
		ProjectID: projectID,
	}
	db := database.NewMemoryDatabase()

	service := NewCaptionsService(&cfg, db)

	assert := assert.New(t)

	assert.NotNil(service)

	assert.NotNil(service.Prefix())
	assert.Equal(service.Prefix(), "")

	assert.NotNil(service.client.Logger)
	if !reflect.DeepEqual(service.client.Logger, logger) {
		t.Errorf("Wrong logger\nExpected: %#v\nGot:  %#v", logger, service.client.Logger)
	}

	assert.NotNil(service.client.Providers)
	assert.Equal(service.client.Providers, providers)

	assert.Len(service.Endpoints(), 3)
	assert.Contains(service.Endpoints(), "/captions/{id}")
	assert.Contains(service.Endpoints(), "/jobs/{id}")
	assert.Contains(service.Endpoints(), "/captions")
}
