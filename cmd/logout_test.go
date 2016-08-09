package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteFileSucceed(t *testing.T) {
	successfullyDeleted, err := deleteFile()
	assert.Equal(t, successfullyDeleted, true)
	assert.Nil(t, err)
}

func TestDeleteFileFail(t *testing.T) {
	successfullyDeleted, err := deleteFile()
	assert.Equal(t, successfullyDeleted, false)
	assert.Contains(t, err.Error(), "no such file or directory")
}
