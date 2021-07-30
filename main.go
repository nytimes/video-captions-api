package main

import (
	"fmt"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/config"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/NYTimes/video-captions-api/service"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	goprom "github.com/prometheus/client_golang/prometheus"
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
	// metrics server

	exporter, registry := mustInitMetrics()
	go func(log *logrus.Entry) {
		addr := ":9000"
		log.WithField("address", addr).Info("starting metric server")

		mux := http.NewServeMux()
		mux.Handle("/metrics", exporter)

		metricsServer := &http.Server{
			Addr:         addr,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
			Handler:      mux,
		}
		var err error
		if err = metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			log.WithFields(logrus.Fields{
				"err":     err,
				"address": addr,
			}).Fatal("Metrics server failure")
		}

	}(server.Log.WithField("service", "metricsServer"))

	// caption server

	threeplayConfig := providers.Load3PlayConfigFromEnv()
	amaraConfig := providers.LoadAmaraConfigFromEnv()
	captionsService := service.NewCaptionsService(&cfg, db, registry)

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

func mustInitMetrics() (*prometheus.Exporter, *goprom.Registry) {
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
		Namespace: MetricsNamespace,
		Name:      "version",
		Help:      "Application version.",
	}, []string{"version"})

	r.MustRegister(versionCollector)
	versionCollector.WithLabelValues(version).Add(1)

	captionTimer := goprom.NewHistogramVec(goprom.HistogramOpts{
		Namespace: MetricsNamespace,
		Name:      "asr_execution_time_seconds",
		Help:      "provider caption time",
		Buckets:   goprom.LinearBuckets(20, 5, 5),
	}, []string{
		"provider",
	})

	r.MustRegister(captionTimer)

	// Stats exporter: Prometheus
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: MetricsNamespace,
		Registry:  r,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create the Prometheus stats exporter %w", err)
	}

	return pe, r, nil
}
