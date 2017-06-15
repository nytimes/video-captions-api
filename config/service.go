package config

import (
	"github.com/NYTimes/gizmo/server"
	log "github.com/Sirupsen/logrus"
)

// CaptionsServiceConfig is the configuration needed to create a CaptionsService
type CaptionsServiceConfig struct {
	Server *server.Config
	Logger *log.Logger
}
