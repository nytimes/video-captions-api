package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NYTimes/video-captions-api/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func makeContext(d time.Duration) (context.Context, func()) {
	return context.WithTimeout(context.Background(), d)
}

func TestCallbackListener(t *testing.T) {
	tests := []struct {
		name    string
		p       []Provider
		ttl     time.Duration
		payload io.Reader
	}{
		{
			name: "Happy Path-single provider",
			p: []Provider{
				&testProvider{
					name: "happy",

					f: func(p *testProvider, req *http.Request) (string, *CallbackData, error) {
						return "TODO", &CallbackData{}, nil
					},
				},
			},
			ttl:     time.Second * 2,
			payload: strings.NewReader("This will surely fail"),
		},
	}

	log := logrus.New()
	for _, tc := range tests {

		//t.Parallel()
		ctx, cancel := makeContext(tc.ttl)
		var wg sync.WaitGroup
		t.Log("starting	callback listener")
		q, uris := StartCallbackListener(
			ctx,
			&wg,
			tc.p,
			log.WithField(
				"testname", tc.name,
			),
		)
		addr := uris[tc.p[0].(*testProvider).name]
		go func() {
			t.Log("calling the callback")
			res, err := http.Post(addr, "application/json", tc.payload)
			require.NoError(t, err)
			fmt.Printf("%+v\n", res)
			cancel()
		}()
		t.Log("ranging over CallbackData q")
		for data := range q {
			// do something
			log.WithField("data", data).Info("Got callback data from q")
		}

		wg.Wait()
	}

}

type testProvider struct {
	name        string
	f           func(*testProvider, *http.Request) (string, *CallbackData, error)
	runningJobs []*database.Job
}

func (p *testProvider) DispatchJob(job *database.Job) error {
	p.runningJobs = append(p.runningJobs, job)

	return nil
}

func (p *testProvider) Download(_ *database.Job, _ string) ([]byte, error) {
	return nil, nil
}

func (p *testProvider) GetProviderJob(_ *database.Job) (*database.ProviderJob, error) {
	return nil, nil
}

func (p *testProvider) GetName() string {
	return p.name
}

func (p *testProvider) HandleCallback(req *http.Request) (string, *CallbackData, error) {
	return p.f(p, req)
}

func (p *testProvider) CancelJob(_ *database.Job) (bool, error) {
	return true, nil
}
