package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/sync/errgroup"

	goprom "github.com/prometheus/client_golang/prometheus"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/NYTimes/gizmo/server"
	videocaptionsapi "github.com/NYTimes/video-captions-api"
	"github.com/NYTimes/video-captions-api/config"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/NYTimes/video-captions-api/service"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)

	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	go func() {
		select {
		case <-interrupt:
			cancel()

		}
	}()
	implementedProviders := makeProviders(&cfg, db)

	exporter, registry := MustInitMetrics()
	videocaptionsapi.StartMetricsServer(ctx, eg, exporter, server.Log)
	callbackQueue, endpoints := providers.StartCallbackListener(ctx, &sync.WaitGroup{}, implementedProviders, server.Log.WithField("service", "CallbackListener"))
	// caption server
	captionsService := service.NewCaptionsService(&cfg, db, callbackQueue, endpoints, registry)

	for _, p := range implementedProviders {
		captionsService.AddProvider(p)
	}
	server.Init("video-captions-api", cfg.Server)

	err = server.Register(captionsService)
	if err != nil {
		server.Log.Fatal("Unable to register service: ", err)
	}

	err = server.Run()
	if err != nil {
		server.Log.Fatal("Server encountered a fatal error: ", err)
	}

	eg.Wait()

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

func MustInitMetrics() (*prometheus.Exporter, *goprom.Registry) {
	pe, r, err := initMetrics()
	if err != nil {
		panic(errors.Wrap(err, "Failed to initialize metrics service"))
	}
	return pe, r
}

func initMetrics() (*prometheus.Exporter, *goprom.Registry, error) {
	r := goprom.NewRegistry()
	r.MustRegister(goprom.NewProcessCollector(goprom.ProcessCollectorOpts{}))
	r.MustRegister(goprom.NewGoCollector())

	versionCollector := goprom.NewGaugeVec(goprom.GaugeOpts{
		Namespace: videocaptionsapi.MetricsNamespace,
		Name:      "version",
		Help:      "Application version.",
	}, []string{"version"})

	r.MustRegister(versionCollector)
	versionCollector.WithLabelValues(version).Add(1)
	// Stats exporter: Prometheus
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: videocaptionsapi.MetricsNamespace,
		Registry:  r,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create the Prometheus stats exporter %w", err)
	}

	return pe, r, nil
}
