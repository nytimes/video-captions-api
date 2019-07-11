package config

import (
	"github.com/NYTimes/gizmo/server"
	log "github.com/sirupsen/logrus"
)

// CaptionsServiceConfig is the configuration needed to create a CaptionsService
type CaptionsServiceConfig struct {
	Server         *server.Config
	ProjectID      string `envconfig:"PROJECT_ID"`
	Logger         *log.Logger
	BucketName     string `envconfig:"BUCKET_NAME"`
	CallbackURL    string `envconfig:"CALLBACK_URL"`
	CallbackAPIKey string `envconfig:"CALLBACK_API_KEY"`
}
