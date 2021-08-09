package videocaptionsapi

import (
	"context"
	"net/http"
	"sync"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func StartMetricsServer(
	ctx context.Context,
	eg *errgroup.Group,
	exporter *prometheus.Exporter,
	log *logrus.Logger) error {

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

	shutdownCtx, cancel := context.WithCancel(ctx)
	eg.Go(func() error {
		defer cancel()
		var err error
		if err = srv.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("server Shutdown Failed:%+s", err)
			return err
		}
		return nil

	})

	eg.Go(func() error {
		log := log.WithField("service", "metrics_server")
		var err error
		if err = srv.ListenAndServe(); err != http.ErrServerClosed {
			log.WithFields(logrus.Fields{
				"err":     err,
				"address": addr,
			}).Fatal("Metrics server failure")
			return err
		}
		return nil

	})
	return eg.Wait()
}

func StartCallbackListener(
	ctx context.Context,
	wg *sync.WaitGroup,
	callers []providers.Provider,
	log *logrus.Logger) chan *providers.DataWrapper {
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
