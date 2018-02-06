package amara

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Language struct {
	ID                     int         `json:"id"`
	Created                time.Time   `json:"created"`
	LanguageCode           string      `json:"language_code"`
	IsPrimaryAudioLanguage bool        `json:"is_primary_audio_language"`
	IsRtl                  bool        `json:"is_rtl"`
	IsTranslation          bool        `json:"is_translation"`
	Published              bool        `json:"published"`
	OriginalLanguageCode   interface{} `json:"original_language_code"`
	Name                   string      `json:"name"`
	Title                  string      `json:"title"`
	Description            string      `json:"description"`
	Metadata               struct {
		SpeakerName string `json:"speaker-name"`
		Location    string `json:"location"`
	} `json:"metadata"`
	SubtitleCount     int  `json:"subtitle_count"`
	SubtitlesComplete bool `json:"subtitles_complete"`
	Versions          []struct {
		Author struct {
			Username string `json:"username"`
			ID       string `json:"id"`
			URI      string `json:"uri"`
		} `json:"author"`
		Published bool `json:"published"`
		VersionNo int  `json:"version_no"`
	} `json:"versions"`
	SubtitlesURI string `json:"subtitles_uri"`
	ResourceURI  string `json:"resource_uri"`
	NumVersions  int    `json:"num_versions"`
	IsOriginal   bool   `json:"is_original"`
}

type Video struct {
	ID                       string      `json:"id"`
	VideoType                string      `json:"video_type"`
	PrimaryAudioLanguageCode string      `json:"primary_audio_language_code"`
	OriginalLanguage         string      `json:"original_language"`
	Title                    string      `json:"title"`
	Description              string      `json:"description"`
	Duration                 int         `json:"duration"`
	Thumbnail                string      `json:"thumbnail"`
	Created                  time.Time   `json:"created"`
	Team                     interface{} `json:"team"`
	TeamType                 interface{} `json:"team_type"`
	Project                  interface{} `json:"project"`
	AllUrls                  []string    `json:"all_urls"`
	Metadata                 struct {
		SpeakerName string `json:"speaker-name"`
		Location    string `json:"location"`
	} `json:"metadata"`
	Languages []struct {
		Code         string `json:"code"`
		Name         string `json:"name"`
		Published    bool   `json:"published"`
		Dir          string `json:"dir"`
		SubtitlesURI string `json:"subtitles_uri"`
		ResourceURI  string `json:"resource_uri"`
	} `json:"languages"`
	ActivityURI          string `json:"activity_uri"`
	UrlsURI              string `json:"urls_uri"`
	SubtitleLanguagesURI string `json:"subtitle_languages_uri"`
	ResourceURI          string `json:"resource_uri"`
}

type Subtitles struct {
	VersionNumber int    `json:"version_number"`
	SubFormat     string `json:"sub_format"`
	Subtitles     string `json:"subtitles"`
	Author        struct {
		Username string `json:"username"`
		ID       string `json:"id"`
		URI      string `json:"uri"`
	} `json:"author"`
	Language struct {
		Code string `json:"code"`
		Name string `json:"name"`
		Dir  string `json:"dir"`
	} `json:"language"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Metadata    struct {
		SpeakerName string `json:"speaker-name"`
		Location    string `json:"location"`
	} `json:"metadata"`
	VideoTitle       string `json:"video_title"`
	VideoDescription string `json:"video_description"`
	ActionsURI       string `json:"actions_uri"`
	NotesURI         string `json:"notes_uri"`
	ResourceURI      string `json:"resource_uri"`
	SiteURI          string `json:"site_uri"`
	Video            string `json:"video"`
	VersionNo        int    `json:"version_no"`
}

type EditorLoginSession struct {
	URL string `json:"url"`
}

func (c *Client) GetVideo(id string) (*Video, error) {
	data, err := c.doRequest("GET", fmt.Sprintf("%s/videos/%s/", c.endpoint, id), nil)
	if err != nil {
		return nil, err
	}
	video := Video{}
	if err = json.Unmarshal(data, &video); err != nil {
		return nil, err
	}
	return &video, nil
}

func (c *Client) CreateVideo(params url.Values) (*Video, error) {
	data, err := c.doRequest(
		"POST",
		fmt.Sprintf("%s/videos/", c.endpoint),
		bytes.NewBufferString(params.Encode()),
	)
	if err != nil {
		return nil, err
	}
	video := Video{}
	if err = json.Unmarshal(data, &video); err != nil {
		return nil, err
	}
	return &video, nil
}

func (c *Client) GetLanguage(videoID, langCode string) (*Language, error) {
	data, err := c.doRequest(
		"GET",
		fmt.Sprintf("%s/videos/%s/languages/%s/", c.endpoint, videoID, langCode),
		nil,
	)
	if err != nil {
		return nil, err
	}
	language := Language{}
	if err = json.Unmarshal(data, &language); err != nil {
		return nil, err
	}
	return &language, nil
}

func (c *Client) CreateLanguage(videoID, langCode string) (*Language, error) {
	params := url.Values{}
	params.Add("language_code", langCode)
	params.Add("subtitles_complete", "false")
	params.Add("is_primary_audio_language", "true")
	data, err := c.doRequest(
		"POST",
		fmt.Sprintf("%s/videos/%s/languages/", c.endpoint, videoID),
		strings.NewReader(params.Encode()),
	)

	if err != nil {
		return nil, err
	}

	lang := Language{}
	if err = json.Unmarshal(data, &lang); err != nil {
		return nil, err
	}
	return &lang, nil
}

func (c *Client) UpdateLanguage(videoID, langCode string, complete bool) (*Language, error) {
	params := url.Values{}
	params.Add("subtitles_complete", strconv.FormatBool(complete))
	data, err := c.doRequest(
		"PUT",
		fmt.Sprintf("%s/videos/%s/languages/%s/", c.endpoint, videoID, langCode),
		strings.NewReader(params.Encode()),
	)

	if err != nil {
		return nil, err
	}

	lang := Language{}
	if err = json.Unmarshal(data, &lang); err != nil {
		return nil, err
	}
	return &lang, nil
}

func (c *Client) CreateSubtitles(videoID, langCode, format string, params url.Values) (*Subtitles, error) {
	if params == nil {
		return nil, errors.New("Please provide the request body parameters")
	}

	params.Set("sub_format", format)
	data, err := c.doRequest(
		"POST",
		fmt.Sprintf("%s/videos/%s/languages/%s/subtitles/", c.endpoint, videoID, langCode),
		bytes.NewBufferString(params.Encode()),
	)
	if err != nil {
		return nil, err
	}
	subtitle := Subtitles{}
	if err = json.Unmarshal(data, &subtitle); err != nil {
		return nil, err
	}
	return &subtitle, nil
}

func (c *Client) GetSubtitles(videoID, langCode string, captionFormat string) (*Subtitles, error) {
	data, err := c.doRequest(
		"GET",
		fmt.Sprintf("%s/videos/%s/languages/%s/subtitles/?sub_format=%s", c.endpoint, videoID, langCode, captionFormat),
		nil,
	)
	if err != nil {
		return nil, err
	}
	subtitle := Subtitles{}
	if err = json.Unmarshal(data, &subtitle); err != nil {
		return nil, err
	}
	return &subtitle, nil
}

func (c *Client) EditorLogin(videoID, langCode, userName string) (*EditorLoginSession, error) {
	params := url.Values{}
	params.Set("video_id", videoID)
	params.Set("user", userName)
	params.Set("language_code", langCode)

	data, err := c.doRequest(
		"POST",
		fmt.Sprintf("%s/editor-login/", c.endpoint),
		bytes.NewBufferString(params.Encode()),
	)
	if err != nil {
		return nil, err
	}

	var editorSession EditorLoginSession
	if err := json.Unmarshal(data, &editorSession); err != nil {
		return nil, err
	}

	return &editorSession, nil
}
