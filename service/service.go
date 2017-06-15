package service

import (
	"net/http"

	"github.com/NYTimes/gziphandler"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
)

// CaptionsService the service responsible to wrapping interactions with Providers
type CaptionsService struct {
	client Client
}

// Config used for configuration/injection of settings for a CaptionService
type Config struct {
	Server             *server.Config
	ThreePlayAPIKey    string `envconfig:"THREE_PLAY_API_KEY"`
	ThreePlayAPISecret string `envconfig:"THREE_PLAY_API_SECRET"`
}

// NewCaptionsService creates a CaptionsService
func NewCaptionsService(cfg *Config, providersArr []providers.Provider, db database.DB) *CaptionsService {
	providersByName := make(map[string]providers.Provider)
	for _, provider := range providersArr {
		providersByName[provider.GetName()] = provider
	}
	return &CaptionsService{
		Client{
			Providers: providersByName,
			DB:        db,
		},
	}
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
		"/jobs/{id}": map[string]http.HandlerFunc{
			"GET": server.JSONToHTTP(s.GetJob).ServeHTTP,
		},
		"/jobs": map[string]http.HandlerFunc{
			"POST": server.JSONToHTTP(s.CreateJob).ServeHTTP,
		},
	}
}
