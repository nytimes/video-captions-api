package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newJob() Job {
	return Job{
		ID:     "123",
		Status: "in_progress",
		Done:   false,
	}
}

func TestJobUpdate(t *testing.T) {
	assert := assert.New(t)
	job := newJob()

	assert.True(job.UpdateStatus("delivered", ""))
	assert.False(job.UpdateStatus("delivered", ""))
	assert.Equal(job.Details, "")

	assert.True(job.UpdateStatus("error", "error details"))
	assert.True(job.Done)
	assert.Equal(job.Details, "error details")
	assert.False(job.UpdateStatus("error", "more details"))
	assert.Equal(job.Details, "error details")
}
