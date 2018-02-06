package threeplay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// TranscriptFormat supported output formats for transcripts
type TranscriptFormat string

const (
	// JSON format for transcripted file
	JSON TranscriptFormat = "json"
	// TXT format output for transcripted file
	TXT TranscriptFormat = "txt"
	// HTML format output for transcripted file
	HTML TranscriptFormat = "html"
)

// Word words
type Word [2]string

// Transcript video transcript
type Transcript struct {
	Words      []Word            `json:"words"`
	Paragraphs []int             `json:"paragraphs"`
	Speakers   map[string]string `json:"speakers"`
}

// GetTranscript get json transcript by file ID
func (c *Client) GetTranscript(fileID uint) (*Transcript, error) {
	response, err := c.GetTranscriptWithFormat(fileID, JSON)
	if err != nil {
		return nil, err
	}

	transcript := &Transcript{}
	err = json.Unmarshal(response, transcript)
	if err != nil {
		return nil, err
	}

	return transcript, nil
}

// GetTranscriptWithFormat get transcript by file ID with supported formats
// current supported formats are json, text and html
func (c *Client) GetTranscriptWithFormat(id uint, format TranscriptFormat) ([]byte, error) {
	endpoint := fmt.Sprintf("https://%s/files/%d/transcript.%s?apikey=%s",
		threePlayStaticHost, id, format, c.apiKey,
	)

	response, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, err
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if err := checkForAPIError(responseData); err != nil {
		return nil, err
	}

	return responseData, nil
}

// GetTranscriptByVideoID get json transcript by video ID
func (c *Client) GetTranscriptByVideoID(videoID string) (*Transcript, error) {

	response, err := c.GetTranscriptByVideoIDWithFormat(videoID, JSON)
	if err != nil {
		return nil, err
	}
	transcript := &Transcript{}
	err = json.Unmarshal(response, transcript)
	if err != nil {
		return nil, err
	}

	return transcript, nil
}

// GetTranscriptByVideoIDWithFormat get transcript by video ID with specific format
// current supported formats are json, text and html
func (c *Client) GetTranscriptByVideoIDWithFormat(id string, format TranscriptFormat) ([]byte, error) {
	endpoint := fmt.Sprintf("https://%s/files/%s/transcript.%s?apikey=%s&usevideoid=1",
		threePlayStaticHost, id, format, c.apiKey,
	)

	response, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, err
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if err := checkForAPIError(responseData); err != nil {
		return nil, err
	}
	return responseData, nil
}
