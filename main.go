package main

import (
	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/server"
	captionsConfig "github.com/NYTimes/video-captions-api/config"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/NYTimes/video-captions-api/service"
)

func main() {
	var cfg captionsConfig.CaptionsServiceConfig
	config.LoadEnvConfig(&cfg)
	db, err := database.NewDatastoreDatabase(cfg.ProjectID)
	if err != nil {
		server.Log.Fatal("Unable to create Datastore client", err)
	}
	providerConfig := providers.Load3PlayConfigFromEnv()
	captionsService := service.NewCaptionsService(&cfg, db)

	captionsService.AddProvider(providers.New3PlayProvider(&providerConfig, &cfg))
	server.Init("video-captions-api", cfg.Server)

	err = server.Register(captionsService)
	if err != nil {
		server.Log.Fatal("Unable to register service: ", err)
	}

	err = server.Run()
	if err != nil {
		server.Log.Fatal("Server encountered a fatal error: ", err)
	}

}
