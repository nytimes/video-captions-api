package service

import (
	"net/http"

	"github.com/NYTimes/gziphandler"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
)

type SimpleService struct {
	client Client
}

type Config struct {
	Server    *server.Config
	APIKey    string
	APISecret string
}

func NewSimpleService(cfg *Config, providersArr []providers.Provider, db database.DB) *SimpleService {
	providersByName := make(map[string]providers.Provider)
	for _, provider := range providersArr {
		providersByName[provider.GetName()] = provider
	}
	return &SimpleService{
		Client{
			Providers: providersByName,
			DB:        db,
		},
	}
}

func (s *SimpleService) Prefix() string {
	return ""
}

func (s *SimpleService) Middleware(h http.Handler) http.Handler {
	return gziphandler.GzipHandler(h)
}

func (s *SimpleService) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/jobs/{id}": map[string]http.HandlerFunc{
			"GET": server.JSONToHTTP(s.GetJob).ServeHTTP,
		},
		"/jobs": map[string]http.HandlerFunc{
			"POST": server.JSONToHTTP(s.CreateJob).ServeHTTP,
		},
	}
}
