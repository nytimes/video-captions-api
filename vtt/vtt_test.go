package vtt

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		inputFile string
		result    error
	}{
		{
			"testdata/sample.vtt",
			nil,
		},
		{
			"testdata/with-bom.vtt",
			nil,
		},
		{
			"testdata/signature-bad-comment.vtt",
			errors.New("[header] invalid signature, whitespace is required before comment [line 1]"),
		},
		{
			"testdata/signature-comment.vtt",
			nil,
		},
		{
			"testdata/signature-space.vtt",
			nil,
		},
		{
			"testdata/signature-tab.vtt",
			nil,
		},
		{
			"testdata/garbage-signature.vtt",
			errors.New("[header] invalid signature, expecting: \"WEBVTT\", got: \"garbage\" [line 1]"),
		},
		{
			"testdata/signature-no-new-line.vtt",
			errors.New("[header] invalid header: an empty new line is required after the header [line 1]"),
		},
		{
			"testdata/empty.vtt",
			errors.New("file is empty"),
		},
		{
			"testdata/with-header.vtt",
			nil,
		},
		{
			"testdata/cue-invalid-timestamp.vtt",
			errors.New("[cue] invalid start timestamp, expecting: \"00:00:00.000\", got: \"00:11.00\" [line 3]"),
		},
		{
			"testdata/no-space-cue-times-arrow.vtt",
			errors.New("[cue] invalid arrow, expecting: \" --> \", got: \"-->\" [line 3]"),
		},
		{
			"testdata/no-space-cue-times-cue-settings.vtt",
			errors.New("[cue] invalid settings, expecting: \" name:value\", got: \"line:40%\" [line 3]"),
		},
		{
			"testdata/garbage-cue.vtt",
			errors.New("[cue] invalid end timestamp, expecting: \"00:00:00.000\", got: \"\" [line 3]"),
		},
		{
			"testdata/many-comments.vtt",
			nil,
		},
		{
			"testdata/style.vtt",
			nil,
		},
		{
			"testdata/style-invalid-css.vtt",
			errors.New("[style] CSS parse error: expected colon in declaration [line 5]"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.inputFile, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.inputFile)

			if err != nil {
				t.Fatalf("Unable to load  %s", tt.inputFile)
			}

			reader := bytes.NewReader(data)
			err = Validate(reader)

			if tt.result != nil {
				assert.Equal(t, tt.result.Error(), err.Error())
			} else {
				assert.Equal(t, tt.result, err)
			}
		})
	}
}
