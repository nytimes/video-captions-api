package service

import (
	"errors"
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

func (p fakeProvider) DispatchJob(job *providers.Job) error {
	p.logger.Info("dispatching job")
	return nil
}

func (p fakeProvider) GetJob(id string) (*providers.Job, error) {
	if p.params["jobError"] {
		return nil, errors.New("oh no")
	}
	if p.params["jobStatus"] {
		return &providers.Job{Status: "My status"}, nil
	}
	p.logger.Info("fetching job", id)
	return &providers.Job{}, nil
}

func (p fakeProvider) GetName() string {
	return "test-provider"
}

type brokenProvider fakeProvider

func (p brokenProvider) GetName() string {
	return "broken-provider"
}

func (p brokenProvider) DispatchJob(job *providers.Job) error {
	return errors.New("provider error")
}

func (p brokenProvider) GetJob(id string) (*providers.Job, error) {
	p.logger.Info("fetching job", id)
	return nil, errors.New("failed to get job")
}

func createCaptionsService() (*CaptionsService, Client) {
	client := Client{
		Providers: make(map[string]providers.Provider),
		DB:        database.NewMemoryDatabase(),
		Logger:    log.New(),
	}
	service := &CaptionsService{
		client: client,
		logger: log.New(),
	}
	return service, client
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
