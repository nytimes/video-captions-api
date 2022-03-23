package main

import (
	"github.com/NYTimes/gizmo/server"
	"github.com/nytimes/video-captions-api/config"
	"github.com/nytimes/video-captions-api/database"
	"github.com/nytimes/video-captions-api/providers"
	"github.com/nytimes/video-captions-api/service"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	var cfg config.CaptionsServiceConfig
	envconfig.Process("", &cfg)
	cfg.Logger = server.Log
	db, err := database.NewDatastoreDatabase(cfg.ProjectID)
	if err != nil {
		server.Log.Fatal("Unable to create Datastore client", err)
	}
	threeplayConfig := providers.Load3PlayConfigFromEnv()
	amaraConfig := providers.LoadAmaraConfigFromEnv()
	captionsService := service.NewCaptionsService(&cfg, db)

	captionsService.AddProvider(providers.New3PlayProvider(&threeplayConfig, &cfg))
	captionsService.AddProvider(providers.NewAmaraProvider(&amaraConfig, &cfg))
	captionsService.AddProvider(providers.NewUploadProvider(&cfg, db))
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
