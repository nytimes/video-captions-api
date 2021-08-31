package providers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/NYTimes/gizmo/server"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type CallbackHandler interface {
	HandleCallback(req *http.Request) (jobID string, data *CallbackData, err error)
}

func StartCallbackListener(
	ctx context.Context,
	port int,
	wg *sync.WaitGroup,
	callbackHandlers []Provider,
	log *logrus.Entry) (chan *DataWrapper, map[string]string) {

	mux := http.NewServeMux()
	q := make(chan *DataWrapper)
	// generate a unique callback URL for each provider
	var uris = make(map[string]string)
	for _, c := range callbackHandlers {

		h, guid := handleRegister(ctx, q, c, log.WithField("provider", c))
		log.WithFields(logrus.Fields{
			"guid": guid,
		}).Info("generating callback urls")
		uris[c.GetName()] = fmt.Sprintf("%s", guid)
		mux.HandleFunc(
			fmt.Sprintf("%s", guid),
			h,
		)
	}

	wg.Add(1)
	go func(log *logrus.Entry) {
		defer wg.Done()

		log.WithField("port", port).Info("Starting callback listener")

		cbServer := &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
			Handler:      mux,
		}
		go func() {
			select {
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*1)
				defer cancel()
				cbServer.Shutdown(shutdownCtx)
			}
		}()
		var err error
		if err = cbServer.ListenAndServe(); err != http.ErrServerClosed {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Fatal("Callback server failure")
		}

	}(server.Log.WithField("service", "callbackListener"))

	return q, uris
}

func handleRegister(ctx context.Context, q chan<- *DataWrapper, c CallbackHandler, log *logrus.Entry) (http.HandlerFunc, string) {
	guid := uuid.NewV4().String()
	path := fmt.Sprintf("/%s", guid)

	h := func(w http.ResponseWriter, req *http.Request) {
		log.Info("handle resgister callback")
		id, data, err := c.HandleCallback(req)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error":    err,
				"endpoint": guid,
			}).Error("Failed to handle callback")
			http.Error(w, "Callback Failed", http.StatusInternalServerError)
		}
		q <- &DataWrapper{
			JobID: id,
			Data:  data,
		}
	}
	return h, path
}
