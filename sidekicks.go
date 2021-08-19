package videocaptionsapi

import (
	"context"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/sirupsen/logrus"

	"golang.org/x/sync/errgroup"
)

const (
	// MetricsNamespace is the name of the application.
	MetricsNamespace = "video_captions_api"
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
