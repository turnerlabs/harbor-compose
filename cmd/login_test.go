package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteFileSuccess(t *testing.T) {
	var version = "v1"
	var username = "testtest"
	var token = "1234567"
	successfullyWritten, err := writeFile(version, username, token)
	assert.Nil(t, err)
	assert.Equal(t, successfullyWritten, true)
}

func TestReadFileSuccess(t *testing.T) {
	auth, err := readFile()
	assert.Nil(t, err)
	assert.Equal(t, auth.Version, "v1")
	assert.Equal(t, auth.Username, "testtest")
	assert.Equal(t, auth.Token, "1234567")
}

func TestHarborLoginFail(t *testing.T) {
	t.SkipNow()
	var usernameTest = "testtest"
	var password = "huksdhjkashdj"
	token, err := harborLogin(usernameTest, password)
	assert.Empty(t, token)
	assert.Equal(t, err.Error(), "Invalid Status Code: 401 Unauthorized")
}

func TestHarborAuthenticatedFail(t *testing.T) {
	t.SkipNow()
	var usernameTest = "testtest"
	var token = "huksdhjkashdj"
	isTokenValid, err := harborAuthenticated(usernameTest, token)
	assert.Equal(t, isTokenValid, false)
	assert.Equal(t, err.Error(), "Invalid Status Code: 401 Unauthorized")
}

func TestHarborLoginSuccess(t *testing.T) {
	if usernameTest == "" || passwordTest == "" {
		t.SkipNow()
	}
	token, err := harborLogin(usernameTest, passwordTest)
	assert.NotEmpty(t, token)
	assert.Nil(t, err)
}

func TestHarborAuthenticatedSuccess(t *testing.T) {
	if usernameTest == "" || passwordTest == "" {
		t.SkipNow()
	}
	token, err := harborLogin(usernameTest, passwordTest)
	assert.NotEmpty(t, token)
	assert.Nil(t, err)
	isTokenValid, err := harborAuthenticated(usernameTest, token)
	assert.Equal(t, isTokenValid, true)
	assert.Nil(t, err)
}
