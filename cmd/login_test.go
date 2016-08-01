package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// uses passed in username and password
func TestLogin(t *testing.T) {
	token, err := Login(username, password)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)
}

// uses passed in username and password
func TestLoginFail(t *testing.T) {
	var badUserName = "test"
	var badPassword = "test123456"

	token, err := Login(badUserName, badPassword)
	assert.Empty(t, token)
	assert.Equal(t, err.Error(), "Invalid Status Code: 401 Unauthorized")
}
