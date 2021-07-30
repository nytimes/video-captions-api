package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/pkg/errors"
	goprom "github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func StartMetricsServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	exporter *prometheus.Exporter,
	log *logrus.Logger) {

	addr := ":9000"
	log.WithField("address", addr).Info("starting metric server")

	mux := http.NewServeMux()
	mux.Handle("/metrics", exporter)

	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		Handler:      mux,
	}
	wg.Add(1)

	shutdownCtx, cancel := context.WithCancel(ctx)
	go func() {
		defer wg.Done()
		defer cancel()
		var err error
		if err = srv.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("server Shutdown Failed:%+s", err)
		}

	}()

	wg.Add(1)
	go func(log *logrus.Entry) {
		defer wg.Done()
		var err error
		if err = srv.ListenAndServe(); err != http.ErrServerClosed {
			log.WithFields(logrus.Fields{
				"err":     err,
				"address": addr,
			}).Fatal("Metrics server failure")
		}
	}(log.WithField("service", "metrics_server"))

}

func StartCallbackListener(
	ctx context.Context,
	wg *sync.WaitGroup,
	callers []providers.Provider,
	log *logrus.Logger) chan *providers.DataWrapper {
	wg.Add(1)
	// callback server
	wg.Add(1)
	callbacks := make(chan *providers.DataWrapper)
	go func(log *logrus.Entry) {
		defer wg.Done()

		mux := http.NewServeMux()

		cbServer := &http.Server{
			Addr:         ":9090",
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
			Handler:      mux,
		}
		var err error
		if err = cbServer.ListenAndServe(); err != http.ErrServerClosed {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Fatal("Callback server failure")
		}

	}(server.Log.WithField("service", "callbackListener"))

	return callbacks
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
