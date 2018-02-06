package threeplay

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// File representation
type File struct {
	ID                   uint   `json:"id"`
	ProjectID            uint   `json:"project_id"`
	BatchID              uint   `json:"batch_id"`
	Duration             uint   `json:"duration"`
	Attribute1           string `json:"attribute1"`
	Attribute2           string `json:"attribute2"`
	Attribute3           string `json:"attribute3"`
	VideoID              string `json:"video_id"`
	Name                 string `json:"name"`
	CallbackURL          string `json:"callback_url"`
	Description          string `json:"description"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
	WordCount            uint   `json:"word_count"`
	ThumbnailURL         string `json:"thumbnail_url"`
	LanguageID           int    `json:"language_id"`
	DefaultServiceTypeID int    `json:"default_service_type_id"`
	Downloaded           bool   `json:"downloaded"`
	State                string `json:"state"`
	TurnaroundLevelID    int    `json:"turnaround_level_id"`
	Deadline             string `json:"deadline"`
	BatchName            string `json:"batch_name"`
	ErrorDescription     string `json:"error_description"`
}

// FilesPage representation
type FilesPage struct {
	Files   []File `json:"files"`
	Summary `json:"summary"`
}

// Summary representation
type Summary struct {
	CurrentPage  json.Number `json:"current_page"`
	PerPage      json.Number `json:"per_page"`
	TotalEntries json.Number `json:"total_entries"`
	TotalPages   json.Number `json:"total_pages"`
}

// UpdateFile updates a File metadata
func (c *Client) UpdateFile(fileID uint, data url.Values) error {
	if data == nil {
		return errors.New("Must specify new data")
	}
	apiURL := c.createURL(fmt.Sprintf("/files/%d", fileID))
	data.Set("apikey", c.apiKey)
	data.Set("api_secret_key", c.apiSecret)
	req, err := c.createRequest(http.MethodPut, apiURL.String(), data)
	if err != nil {
		return err
	}
	response, err := c.httpClient.Do(req)
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return checkForAPIError(responseData)
}

// GetFiles returns a list of files, supports pagination through params
// and filters.
// For a full list of supported filtering parameters check http://support.3playmedia.com/hc/en-us/articles/227729828-Files-API-Methods
func (c *Client) GetFiles(params, filters url.Values) (*FilesPage, error) {
	querystring := url.Values{}
	if params != nil {
		querystring = params
	}
	if filters != nil {
		querystring.Set("q", filters.Encode())
	}
	filesPage := &FilesPage{}
	url := c.createURL("/files")
	endpoint := c.prepareURL(url, querystring)
	res, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	if err := parseResponse(res, filesPage); err != nil {
		return nil, err
	}
	return filesPage, nil
}

// GetFile gets a single file by id
func (c *Client) GetFile(id uint) (*File, error) {
	file := &File{}
	url := c.createURL(fmt.Sprintf("/files/%d", id))
	endpoint := c.prepareURL(url, nil)
	res, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	if err := parseResponse(res, file); err != nil {
		return nil, err
	}
	return file, nil
}

// UploadFileFromURL uploads a file to threeplay using the file's URL and
// returns the file ID.
func (c *Client) UploadFileFromURL(fileURL string, options url.Values) (uint, error) {
	apiURL := c.createURL("/files")
	data := url.Values{}
	data.Set("apikey", c.apiKey)
	data.Set("api_secret_key", c.apiSecret)
	data.Set("link", fileURL)
	for key, val := range options {
		data[key] = val
	}
	res, err := c.httpClient.PostForm(apiURL.String(), data)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	err = checkForAPIError(responseData)
	if err != nil {
		return 0, err
	}
	fileID, err := strconv.ParseUint(string(responseData), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid response: %s", responseData)
	}

	return uint(fileID), nil
}
