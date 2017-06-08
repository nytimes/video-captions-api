package service

import (
	"net/http"

	"github.com/NYTimes/gziphandler"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/NYTimes/video-captions-api/providers/threeplay"
)

type SimpleService struct {
	client providers.Provider
}

type Config struct {
	Server    *server.Config
	APIKey    string
	APISecret string
}

func NewSimpleService(cfg *Config) *SimpleService {
	return &SimpleService{
		threeplay.New(cfg.APIKey, cfg.APISecret),
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
