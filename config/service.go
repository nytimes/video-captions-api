package config

import (
	"github.com/NYTimes/gizmo/server"
	log "github.com/Sirupsen/logrus"
)

// CaptionsServiceConfig is the configuration needed to create a CaptionsService
type CaptionsServiceConfig struct {
	Server     *server.Config
	ProjectID  string `envconfig:"PROJECT_ID"`
	Logger     *log.Logger
	BucketName string `envconfig:"BUCKET_NAME"`
}
