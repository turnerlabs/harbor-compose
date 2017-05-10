package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHarborLoginFail(t *testing.T) {
	if !*integrationTest || *usernameTest == "" || *passwordTest == "" {
		t.SkipNow()
	}
	var usernameTest = "testtest"
	var password = "huksdhjkashdj"
	token, err := harborLogin(usernameTest, password)
	assert.Empty(t, token)
	assert.Equal(t, err.Error(), "Invalid Status Code: 401 Unauthorized")
}

func TestHarborAuthenticatedFail(t *testing.T) {
	if !*integrationTest || *usernameTest == "" || *passwordTest == "" {
		t.SkipNow()
	}
	var usernameTest = "testtest"
	var token = "huksdhjkashdj"
	isTokenValid, err := harborAuthenticated(usernameTest, token)
	assert.Equal(t, isTokenValid, false)
	assert.Equal(t, err.Error(), "Invalid Status Code: 401 Unauthorized")
}

func TestHarborLoginSuccess(t *testing.T) {
	if !*integrationTest || *usernameTest == "" || *passwordTest == "" {
		t.SkipNow()
	}
	token, err := harborLogin(*usernameTest, *passwordTest)
	assert.NotEmpty(t, token)
	assert.Nil(t, err)
}

func TestHarborAuthenticatedSuccess(t *testing.T) {
	if !*integrationTest || *usernameTest == "" || *passwordTest == "" {
		t.SkipNow()
	}
	token, err := harborLogin(*usernameTest, *passwordTest)
	assert.NotEmpty(t, token)
	assert.Nil(t, err)
	isTokenValid, err := harborAuthenticated(*usernameTest, token)
	assert.Equal(t, isTokenValid, true)
	assert.Nil(t, err)
}
