package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	videocaptionsapi "github.com/NYTimes/video-captions-api"
	"github.com/NYTimes/video-captions-api/database"
	"github.com/NYTimes/video-captions-api/providers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var captionTimer = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: videocaptionsapi.MetricsNamespace,
	Name:      "asr_execution_time_seconds",
	Help:      "provider caption time",
	Buckets:   prometheus.LinearBuckets(10, 10, 10),
}, []string{
	"provider",
	"job",
})

// Client CaptionsService client
type Client struct {
	Providers      map[string]providers.Provider
	DB             database.DB
	Logger         *log.Logger
	Storage        Storage
	CallbackURL    string
	CallbackAPIKey string
	Metrics        *prometheus.Registry
}

func NewClient(db database.DB, logger *logrus.Logger, storage Storage, metrics *prometheus.Registry) Client {
	metrics.MustRegister(captionTimer)
	return Client{
		Providers: make(map[string]providers.Provider),
		DB:        db,
		Logger:    logger,
		Storage:   storage,
		Metrics:   metrics,
	}

}

// GetJobs gets all jobs associated with a ParentID
func (c Client) GetJobs(parentID string) ([]*database.JobSummary, error) {
	jobs, err := c.DB.GetJobs(parentID)
	if err != nil {
		c.Logger.Errorf("Error loading jobs from DB for parent ID %s: %v", parentID, err)
		return nil, err
	}
	sort.Sort(database.ByCreatedAt(jobs))
	summaries := make([]*database.JobSummary, len(jobs))
	for i, job := range jobs {
		summaries[i] = &database.JobSummary{ID: job.ID, CreatedAt: job.CreatedAt}
	}
	return summaries, nil
}

// GetJob gets a job by ID
func (c Client) GetJob(jobID string) (*database.Job, error) {
	job, err := c.DB.GetJob(jobID)
	if err != nil {
		c.Logger.Error("Could not find Job in database")
		return nil, err
	}

	if job.Done {
		return job, nil
	}

	providerID := job.GetProviderID()
	fields := log.Fields{"JobID": jobID, "Provider": job.Provider, "ProviderID": providerID}
	jobLogger := c.Logger.WithFields(fields)
	provider := c.Providers[job.Provider]
	jobLogger.Info("Fetching job from Provider")
	providerJob, err := provider.GetProviderJob(job)
	if err != nil {
		jobLogger.Error("error getting job from provider", err)
		return nil, err
	}

	params := providerJob.Params

	shouldUpdate := false

	for k, v := range params {
		if params[k] != job.ProviderParams[k] {
			job.ProviderParams[k] = v
			shouldUpdate = true
		}
	}

	if job.UpdateStatus(providerJob.Status, providerJob.Details) || shouldUpdate {
		err = c.DB.UpdateJob(jobID, job)
	}

	if (job.Status == "complete" || job.Status == "delivered") && !job.Done {
		jobLogger.Info("Job is ready on the provider, downloading")
		for i, output := range job.Outputs {
			data, err := provider.Download(job, output.Type)
			if err != nil {
				jobLogger.WithError(err).Error("Failed to download file")
				return job, nil
			}
			jobLogger.Info("Download done, storing")
			dest, err := c.Storage.Store(data, fmt.Sprintf("%s/%s", job.Provider, output.Filename))
			if err != nil {
				jobLogger.WithError(err).Error("Failed to store file")
				return job, nil
			}
			job.Outputs[i].URL = dest
		}
		job.Done = true
		err = c.DB.UpdateJob(jobID, job)
	}
	return job, err
}

// DispatchJob dispatches a Job given an existing Provider
func (c Client) DispatchJob(job *database.Job) error {

	provider := c.Providers[job.Provider]
	jobLogger := c.Logger.WithFields(log.Fields{"JobID": job.ID, "Provider": job.Provider})
	if provider == nil {
		jobLogger.Error("provider not found")
		return errors.New("provider not found")
	}

	jobLogger.Info("Dispatching job to provider")
	err := provider.DispatchJob(job)
	if err != nil {
		jobLogger.Errorf("Error dispatching job to provider: %v", err)
		return fmt.Errorf("Error dispatching Job: %v", err)
	}
	jobLogger.Info("Storing job in DB")
	_, err = c.DB.StoreJob(job)
	if err != nil {
		jobLogger.Errorf("Error storing job in DB: %v", err)
		return fmt.Errorf("Error storing Job: %v", err)
	}
	return nil
}

// CancelJob cancels a job by ID
func (c Client) CancelJob(jobID string) (bool, error) {
	job, err := c.DB.GetJob(jobID)
	if err != nil {
		c.Logger.Error("Could not find Job in database")
		return false, err
	}

	if job.Done {
		c.Logger.Error("Cannot cancel a job that is already done")
		return false, nil
	}

	job.Status = "cancelled"
	job.Done = true

	err = c.DB.UpdateJob(jobID, job)
	c.Logger.Info("Cancelled job in the database")
	if job.Provider == "3play" {
		cancellable, err := c.Providers[job.Provider].CancelJob(job)
		if err != nil {
			if cancellable {
				c.Logger.Errorf("Could not cancel job with 3play but set to cancel in DB: %v", err)
				return true, fmt.Errorf("could not cancel job with 3play but set to cancel in DB: %v", err)
			}
			c.Logger.Error("job is no longer cancellable with 3play but was updated in the DB")
			return true, errors.New("job is no longer cancellable with 3play but was updated in the DB")
		}
	}
	c.Logger.Infof("Cancelled job with %s", job.Provider)
	return true, err
}

// DownloadCaption downloads a caption of a given job in the specified format
func (c Client) DownloadCaption(jobID string, captionType string) ([]byte, error) {
	job, err := c.DB.GetJob(jobID)
	if err != nil {
		c.Logger.Error("Could not find Job in database")
		return nil, err
	}

	captionTimer.With(
		prometheus.Labels{
			"provider": job.Provider,
			"job":      job.ID,
			"stage":    "download",
		}).Observe(float64(time.Now().Sub(job.CreatedAt)) / float64(time.Millisecond))

	providerID := job.GetProviderID()
	fields := log.Fields{"JobID": jobID, "Provider": job.Provider, "ProviderID": providerID}
	jobLogger := c.Logger.WithFields(fields)
	provider := c.Providers[job.Provider]
	jobLogger.Info("Downloading captions from provider")
	captions, err := provider.Download(job, captionType)
	if err != nil {
		jobLogger.Error("error downloading captions from provider", err)
		return nil, err
	}
	return captions, nil
}

// GenerateTranscript generates a transcript from the provided caption file and format
func (c Client) GenerateTranscript(captionFile []byte, captionFormat string) (string, error) {
	fields := log.Fields{"captionFormat": captionFormat}
	jobLogger := c.Logger.WithFields(fields)
	jobLogger.Info("Generating transcript for captions")

	type SubtitleParsePreset struct {
		delimiter     string
		linesToIgnore int
		remove        string
		startingIndex int
		splitN        int
	}

	vttPreset := SubtitleParsePreset{
		delimiter:     "\n\n",
		linesToIgnore: 1,
		remove:        "",
		startingIndex: 0,
		splitN:        0,
	}

	srtPreset := SubtitleParsePreset{
		delimiter:     "\r\n\r\n",
		linesToIgnore: 2,
		remove:        "",
		startingIndex: 0,
		splitN:        0,
	}

	sbvPreset := SubtitleParsePreset{
		delimiter:     "\r\n\r\n",
		linesToIgnore: 1,
		remove:        "[br]",
		startingIndex: 0,
		splitN:        0,
	}

	ssaPreset := SubtitleParsePreset{
		delimiter:     "\n",
		linesToIgnore: 0,
		remove:        "",
		startingIndex: 4,
		splitN:        10,
	}

	var parsingPresets = make(map[string]SubtitleParsePreset)
	parsingPresets["vtt"] = vttPreset
	parsingPresets["srt"] = srtPreset
	parsingPresets["sbv"] = sbvPreset
	parsingPresets["ssa"] = ssaPreset

	if _, ok := parsingPresets[captionFormat]; ok {
		subtitleFile := string(captionFile)
		subtitleBlobs := strings.Split(subtitleFile, parsingPresets[captionFormat].delimiter)
		transcript := []string{}

		for i := parsingPresets[captionFormat].startingIndex; i < len(subtitleBlobs); i++ {
			currentBlob := subtitleBlobs[i]
			if parsingPresets[captionFormat].splitN != 0 {
				blobLines := strings.SplitN(currentBlob, ",", 10)
				transcript = append(transcript, strings.TrimSpace(blobLines[len(blobLines)-1]))
			} else {
				blobLines := strings.Split(currentBlob, "\n")
				for j := parsingPresets[captionFormat].linesToIgnore; j < len(blobLines); j++ {
					if len(blobLines[j]) > 0 {
						if parsingPresets[captionFormat].remove != "" {
							cleanString := strings.Replace(blobLines[j], parsingPresets[captionFormat].remove, " ", -1)
							transcript = append(transcript, strings.TrimSpace(cleanString))
						} else {
							transcript = append(transcript, strings.TrimSpace(blobLines[j]))
						}
					}
				}
			}
		}
		return strings.Join(transcript, " "), nil
	}
	jobLogger.Error("error generating a transcript")
	return "", fmt.Errorf("unable to generate a transcript for caption format: %v", captionFormat)
}

func (c Client) notify(jobID string, providerID int, log *logrus.Entry) error {

	log.Debug("Processing a callback for captions")
	if jobID == "" {
		databaseJob, err := c.DB.GetJobByProviderID(strconv.Itoa(providerID))
		if err != nil {
			// We don't need to add fields for jobID or providerID. They are already in the *logrus.Entry. Tell your friends. (The errors below)
			log.WithField("error", err).Error("Failed to get job by provider ID")
			return err
		}
		jobID = databaseJob.ID
	}
	job, err := c.GetJob(jobID)
	if err != nil {
		log.WithField("error", err).Error("Failed to get job by ID")
		return err
	}

	b, err := json.Marshal(job)
	if err != nil {
		log.WithField("error", err).Error("Failed to marshal user response")
		return err
	}

	log.WithField("addr", c.CallbackURL).Debug("Calling API")

	resp, err := http.Post(c.CallbackURL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.WithField("error", err).Error("Failed to POST job completion")
		return err
	}

	defer resp.Body.Close()
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		entry := log.WithFields(logrus.Fields{
			"status": resp.Status,
		})

		msg := "Unexpected response code from user callback"

		_, err := io.ReadAll(resp.Body)
		if err != nil {
			err := fmt.Errorf("%s:%w", msg, err)
			entry.WithField("error", err).Error("failed to read response body")
			return err

		}

		entry.Error(msg)
		return errors.New(msg)
	}

	return nil
}
func (c Client) ProcessCallback(callbackData *providers.CallbackData, jobID string) {

	entry := c.Logger.WithFields(log.Fields{
		"JobID":      jobID,
		"ProviderID": callbackData.ID,
	})

	c.notify(jobID, callbackData.ID, entry)
}
