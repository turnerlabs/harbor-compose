package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteFileSucceed(t *testing.T) {
	//write a file so we can test deleting it
	success, err := writeFile("v1", "foo", "foo")
	assert.Nil(t, err)
	assert.Equal(t, true, success)

	success, err = deleteFile()
	assert.Nil(t, err)
	assert.Equal(t, true, success)
}
