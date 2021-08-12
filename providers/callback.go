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
	wg *sync.WaitGroup,
	callbackHandlers []Provider,
	log *logrus.Entry) (chan *DataWrapper, map[string]string) {

	mux := http.NewServeMux()
	q := make(chan *DataWrapper)
	// generate a unique callback URL for each provider
	var uris = make(map[string]string)
	for _, c := range callbackHandlers {

		h, guid := handleRegister(ctx, q, c, log.WithField("provider", c))
		mux.HandleFunc(
			fmt.Sprintf("%s", guid),
			h,
		)
	}

	wg.Add(1)
	go func(log *logrus.Entry) {
		defer wg.Done()

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

	return q, uris
}

func handleRegister(ctx context.Context, q chan<- *DataWrapper, c CallbackHandler, log *logrus.Entry) (http.HandlerFunc, string) {
	guid := uuid.NewV4().String()
	path := fmt.Sprintf("/%s", guid)

	h := func(w http.ResponseWriter, req *http.Request) {
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
