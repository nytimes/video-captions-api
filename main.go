package main

import (
	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/config"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/NYTimes/video-captions-api/service"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		db  database.DB
		err error
	)
	var cfg config.CaptionsServiceConfig
	envconfig.Process("", &cfg)
	cfg.Logger = server.Log
	if err != nil {
		server.Log.WithFields(
			logrus.Fields{
				"err": err,
			},
		).Info("Invalid value for linker flag main.inmemorydbpermittted. Using remote store provided in PROJECT_ID and BUCKET_NAME")

	}
	server.Log.WithFields(
		logrus.Fields{
			"projectid":  cfg.ProjectID,
			"bucketname": cfg.BucketName,
			"callback":   cfg.CallbackURL,
		},
	).Info("Server Starting")
	if cfg.ProjectID != "" {
		db, err = database.NewDatastoreDatabase(cfg.ProjectID)
		if err != nil {
			server.Log.Fatal("Unable to create Datastore client", err)
		}
	} else {
		server.Log.Warn("Project ID is empty. Using in memory datastore. Are you sure this is what you want?")
		db = database.NewMemoryDatabase()

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
