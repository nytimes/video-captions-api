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

func StartCallbackListener(
	ctx context.Context,
	wg *sync.WaitGroup,
	callers []Provider,
	log *logrus.Logger) (chan *DataWrapper, map[string]string) {

	mux := http.NewServeMux()
	// generate a unique callback URL for each provider
	var uris = make(map[string]string)
	for _, c := range callers {

		guid := uuid.NewV4().String()
		uris[c.GetName()] = guid
		expectedProvider := c
		mux.HandleFunc(
			fmt.Sprintf("%s", guid),
			func(w http.ResponseWriter, req *http.Request) {
				_, _ = expectedProvider.HandleCallback(req)

			})

	}
	wg.Add(1)
	callbacks := make(chan *DataWrapper)
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

	return callbacks, uris
}
