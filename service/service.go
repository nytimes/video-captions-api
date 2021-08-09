package service

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/gziphandler"
	"github.com/NYTimes/video-captions-api/config"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// CaptionsService the service responsible to wrapping interactions with Providers
type CaptionsService struct {
	client    Client
	logger    *log.Logger
	callbacks chan *providers.DataWrapper
	metrics   *prometheus.Registry
}

// NewCaptionsService creates a CaptionsService
func NewCaptionsService(
	cfg *config.CaptionsServiceConfig,
	db database.DB,
	callbacks chan *providers.DataWrapper,
	callbackEndpoints map[string]string,
	metrics *prom.Registry,
) *CaptionsService {
	storage, _ := NewGCSStorage(cfg.BucketName, cfg.Logger)
	client := Client{
		Providers: make(map[string]providers.Provider),
		DB:        db,
		Logger:    cfg.Logger,
		Storage:   storage,
	}
	service := &CaptionsService{
		client,
		cfg.Logger,
		callbacks,
		metrics,
	}
	go func(log *logrus.Entry) {
		for wrapper := range callbacks {
			data, id := wrapper.Data, wrapper.JobID
			client.ProcessCallback(data, id)

			// TODO retry

		}

	}(log.WithField("service", "Callback Listener Worker"))
	return service
}

// AddProvider adds a Provider to the CaptionsService
func (s *CaptionsService) AddProvider(provider providers.Provider) {
	s.client.Providers[provider.GetName()] = provider
}

// Prefix CaptionsService API prefix
func (s *CaptionsService) Prefix() string {
	return ""
}

// Middleware gizmo middleware hook
func (s *CaptionsService) Middleware(h http.Handler) http.Handler {
	return gziphandler.GzipHandler(server.CORSHandler(h, ""))
}

// Endpoints returns CaptionsService API endpoints
func (s *CaptionsService) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/captions/{id}": {
			"GET": server.JSONToHTTP(s.GetJobs).ServeHTTP,
		},
		"/jobs/{id}": {
			"GET": server.JSONToHTTP(s.GetJob).ServeHTTP,
		},
		"/captions": {
			"POST": server.JSONToHTTP(s.CreateJob).ServeHTTP,
		},
		"/jobs/{id}/cancel": {
			"POST": server.JSONToHTTP(s.CancelJob).ServeHTTP,
		},
		"/jobs/{id}/download/{captionFormat}": {
			"GET": s.DownloadCaption,
		},
		"/jobs/{id}/transcript/{captionFormat}": {
			"GET": s.GetTranscript,
		},
	}
}
