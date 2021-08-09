package providers

import (
	"net/http"

	"github.com/NYTimes/video-captions-api/database"
)

type Callback struct {
	Code int
	Data CallbackData
}

type DataWrapper struct {
	JobID string
	Data  CallbackData
	URL   string
}

type CallbackData struct {
	Cancellable         bool    `json:"cancellable"`
	Default             bool    `json:"default"`
	ID                  int     `json:"id"`
	BatchID             int     `json:"batch_id"`
	LanguageID          int     `json:"language_id"`
	MediaFileID         int     `json:"media_file_id"`
	LanguageIDs         []int   `json:"language_ids"`
	CancellationDetails string  `json:"cancellation_details"`
	CancellationReason  string  `json:"cancellation_reason"`
	ReferenceID         string  `json:"reference_id"`
	Status              string  `json:"status"`
	Type                string  `json:"type"`
	Duration            float64 `json:"duration"`
}

// Provider is the interface that transcription/captions providers must implement
type Provider interface {
	DispatchJob(*database.Job) error
	Download(*database.Job, string) ([]byte, error)
	GetProviderJob(*database.Job) (*database.ProviderJob, error)
	GetName() string
	HandleCallback(req *http.Request) (*CallbackData, error)
	CancelJob(*database.Job) (bool, error)
}
