package service

import (
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/config"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	log "github.com/Sirupsen/logrus"
	"github.com/knq/sdhook"
)

var logger = log.New()

func init() {
	if os.Getenv("ENABLE_STACKDRIVER") == "true" {
		logger.Info("Initializing Stackdriver hook")
		stackdriver, err := sdhook.New(
			sdhook.GoogleServiceAccountCredentialsFile("/bin/gcloud.json")
			sdhook.ErrorReportingService("video-captions-api"),
			sdhook.LogName("video-captions-api"),
		)
		if err != nil {
			logger.Error("Failed to initialize Stackdriver hook: ", err)
			return
		}
		logger.Hooks.Add(stackdriver)
	}
}

// CaptionsService the service responsible to wrapping interactions with Providers
type CaptionsService struct {
	client Client
}

// NewCaptionsService creates a CaptionsService
func NewCaptionsService(cfg *config.CaptionsServiceConfig, db database.DB) *CaptionsService {
	if cfg.Logger == nil {
		cfg.Logger = logger
	}
	logger.Info("healthcheck", cfg.Server.HealthCheckPath)
	return &CaptionsService{
		Client{
			Providers: make(map[string]providers.Provider),
			DB:        db,
			logger:    logger,
		},
	}
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
	return gziphandler.GzipHandler(h)
}

// Endpoints returns CaptionsService API endpoints
func (s *CaptionsService) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/jobs/{id}": {
			"GET": server.JSONToHTTP(s.GetJob).ServeHTTP,
		},
		"/jobs": {
			"POST": server.JSONToHTTP(s.CreateJob).ServeHTTP,
		},
	}
}
