package service

import (
	"errors"
	"testing"

	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type fakeProvider struct {
	logger *log.Logger
}

func (p fakeProvider) DispatchJob(job providers.Job) (providers.Job, error) {
	p.logger.Info("dispatching job")
	return job, nil
}

func (p fakeProvider) GetJob(id string) (*providers.Job, error) {
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

func (p brokenProvider) DispatchJob(job providers.Job) (providers.Job, error) {
	return providers.Job{}, errors.New("provider error")
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
