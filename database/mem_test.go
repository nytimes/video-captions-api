package database

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	log = logrus.New()
)

func TestMemStore(t *testing.T) {
	tests := []struct {
		id   string
		job  Job
		name string
	}{
		{
			job:  Job{ID: "ID"},
			name: "happy path",
		},
	}
	db := NewMemoryDatabase()

	for _, tc := range tests {
		id, err := db.StoreJob(&tc.job)
		require.NoError(t, err, "", "Error in  test %s", tc.name)
		require.Equal(t, id, tc.job.ID, "Wrong ID in test %s", tc.name)
	}

}

func TestGetJobID(t *testing.T) {
	tests := []struct {
		job  Job
		name string
	}{
		{
			job: Job{
				ID:       "getjobfakeid",
				Language: "en",
				Done:     true,
			},
			name: "happy path",
		},
	}
	db := NewMemoryDatabase()

	for _, tc := range tests {
		id, err := db.StoreJob(&tc.job)
		require.NoError(t, err, "", "Error in  test %s", tc.name)

		job, err := db.GetJob(id)
		require.NoError(t, err)

		require.Equal(t, job.ID, tc.job.ID)
		require.Equal(t, job.Language, tc.job.Language)
		require.Equal(t, job.Done, tc.job.Done)
	}

}

func generateJobs(t *testing.T, n int) *sync.Map {
	var jobs sync.Map
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("job-%d", i)
		job := &Job{
			ID: id,
		}
		jobs.Store(id, job)

	}
	return &jobs
}

func TestStoreAndGetParallel(t *testing.T) {

	tests := []struct {
		count       int
		name        string
		concurrency int
	}{
		{
			count:       100,
			name:        "parallel happy path",
			concurrency: 10,
		},
	}

	for _, tc := range tests {
		db := NewMemoryDatabase()
		ctx, cancel := context.WithCancel(context.Background())

		jobs := generateJobs(t, tc.count)
		pLog := log.WithField("producer", 0)
		toValidate := jobQueue(ctx, t, jobs, db, pLog)
		// I know this seems way more complicated that it should be. Why not just wait and check at the end?
		// The reason is that we want to hit the store with as many requests as possible as fast as possible to
		// make sure the lock is working as designed.
		var wg sync.WaitGroup
		var runningTotal uint64
		for c := 0; c < tc.concurrency; c++ {
			wg.Add(1)
			log.Debug(
				"starting worker",
				"name", c,
				"population", tc.concurrency,
			)
			wLog := logrus.WithField("worker", c)
			go func(name string, count int, log *logrus.Entry) {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						log.Info("ctx.Done")
						total := atomic.LoadUint64(&runningTotal)
						log.Debug(
							"Context completed. Attempting to validate completion count",
							"total", total,
						)
						if total != uint64(tc.count) {
							t.Errorf("Test %s timed out: error %v", tc.name, ctx.Err())
						}
						cancel()
						return

					case j := <-toValidate:
						log.Info("toValidate")
						total := atomic.AddUint64(&runningTotal, 1)
						log.Debug(
							"pulling job",
							"ID", j.ID,
							"total", total,
						)

						got, err := db.GetJob(j.ID)
						require.NoError(t, err, "Error in %s", tc.name)
						log.Debug(
							"retreived job from store",
							"key", j.ID,
							"id", got.ID,
						)

						e, ok := jobs.Load(got.ID)
						require.True(t, ok)

						expect := e.(*Job).ID
						require.Equal(t, expect, got.ID)

						if total == uint64(tc.count) {
							// Success
							cancel()
							return
						}
					}

				}
			}(tc.name, tc.count, wLog)
		}
		log.Info("waiting")
		wg.Wait()
		cancel()
	}
}

func jobQueue(ctx context.Context, t *testing.T, jobs *sync.Map, db DB, log *logrus.Entry) <-chan *Job {
	q := make(chan *Job)

	go func() {
		jobs.Range(func(id, j interface{}) bool {
			job := j.(*Job)
			select {
			case <-ctx.Done():
				err := ctx.Err()
				t.Errorf("Test timed out: err= %v", err)
				return false
			default:
				db.StoreJob(job)
				log.Debug(
					"pushing job",
					"id", job.ID,
				)
				q <- job
			}
			return true
		})
	}()

	return q

}
