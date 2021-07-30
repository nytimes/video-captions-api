package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/config"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/NYTimes/video-captions-api/service"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

const (
	// MetricsNamespace is the name of the application.
	MetricsNamespace = "video_captions_api"
)

var (
	version = "no version from LDFLAGS"
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

	interrupt := make(chan os.Signal, 1)

	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	var wg sync.WaitGroup
	implementedProviders := makeProviders(&cfg, db)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exporter, registry := MustInitMetrics()
	StartMetricsServer(ctx, &wg, exporter, server.Log)

	callbacks := StartCallbackListener(ctx, &wg, implementedProviders, server.Log)

	// caption server
	captionsService := service.NewCaptionsService(&cfg, db, callbacks, registry)

	for _, p := range implementedProviders {
		captionsService.AddProvider(p)
	}
	server.Init("video-captions-api", cfg.Server)

	//server.WithCloseHandler TODO need to implement gizmo server.Context.Handler to gracefully shut down

	err = server.Register(captionsService)
	if err != nil {
		server.Log.Fatal("Unable to register service: ", err)
	}

	err = server.Run()
	if err != nil {
		server.Log.Fatal("Server encountered a fatal error: ", err)
	}

	wg.Wait()

}

func makeProviders(cfg *config.CaptionsServiceConfig, db database.DB) []providers.Provider {
	var p []providers.Provider
	threeplayConfig := providers.Load3PlayConfigFromEnv()
	amaraConfig := providers.LoadAmaraConfigFromEnv()
	providers.New3PlayProvider(&threeplayConfig, cfg)
	providers.NewAmaraProvider(&amaraConfig, cfg)
	providers.NewUploadProvider(cfg, db)

	return p

}
