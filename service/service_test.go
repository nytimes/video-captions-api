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
	log "github.com/sirupsen/logrus"
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

func (p fakeProvider) Download(_ string, captionType string) ([]byte, error) {
	switch captionType {
	case "vtt":
		return []byte("WEBVTT\n\nNOTE Paragraph\n\n00:00:09.240 --> 00:00:11.010\nWe're all talking\nabout the Iowa caucuses"), nil
	case "srt":
		return []byte("1\r\n00:00:09,240 --> 00:00:11,010\r\nWe’re all talking\r\nabout the Iowa caucuses\r\n\r\n2\n00:00:11,010 --> 00:00:14,180\r\nright now, less than two\r\nweeks till the Iowa caucuses."), nil
	case "sbv":
		return []byte("0:00:09.240,0:00:11.010\nWe're all talking[br]about the Iowa caucuses\r\n\r\n0:00:11.010,0:00:14.190\nright now, less than two[br]weeks till the Iowa caucuses."), nil
	case "ssa":
		return []byte("[Script Info]\nTitle:\n[Events]\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\nDialogue: 0,0:00:00.00,0:00:01.55,Default,,0000,0000,0000,,Some more of the speech"), nil
	default:
		return []byte(""), nil
	}
}

func (p fakeProvider) GetProviderJob(id string) (*database.ProviderJob, error) {
	if p.params["jobError"] {
		return nil, errors.New("oh no")
	}
	if p.params["jobStatus"] {
		return &database.ProviderJob{Status: "My status"}, nil
	}
	if p.params["jobDone"] {
		job := &database.ProviderJob{
			ID:     "123",
			Status: "delivered",
		}
		return job, nil
	}
	p.logger.Info("fetching job", id)
	return &database.ProviderJob{}, nil
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

func (p brokenProvider) GetProviderJob(id string) (*database.ProviderJob, error) {
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

	assert.Contains(service.Endpoints(), "/captions/{id}")
	assert.Contains(service.Endpoints(), "/jobs/{id}")
	assert.Contains(service.Endpoints(), "/captions")
	assert.Contains(service.Endpoints(), "/jobs/{id}/cancel")
	assert.Contains(service.Endpoints(), "/jobs/{id}/download/{captionFormat}")
	assert.Contains(service.Endpoints(), "/jobs/{id}/transcript/{captionFormat}")
}
