package harborauth

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const baduser = "test"
const badpassword = "testtest1234"

var username string
var password string
var url string
var token string

// Common

func GetANil() *string {
	return nil
}

// Flags for testing

func TestMain(m *testing.M) {
	flag.StringVar(&url, "url", "", "url for authorization")
	flag.StringVar(&username, "username", "", "username for authorization")
	flag.StringVar(&password, "password", "", "password for authorization")
	flag.Parse()

	if len(url) == 0 {
		fmt.Println("Missing url.")
		os.Exit(0)
	}

	if len(username) == 0 {
		fmt.Println("Missing username.")
		os.Exit(0)
	}

	if len(password) == 0 {
		fmt.Println("Missing password.")
		os.Exit(0)
	}

	os.Exit(m.Run())
}

// Login

func TestLoginFailBadUsernameAndPassword(t *testing.T) {
	client, err := NewAuthClient(url)
	tokenOut, successOut, err := client.Login(baduser, badpassword)

	assert.Len(t, tokenOut, 0)
	assert.Equal(t, successOut, false)
	assert.Equal(t, "Invalid Status Code: 401 Unauthorized", err.Error())
}

func TestLoginFailUnderMinPassword(t *testing.T) {
	client, err := NewAuthClient(url)
	badPasswordIn := "1234"
	tokenOut, successOut, err := client.Login(baduser, badPasswordIn)

	assert.Len(t, tokenOut, 0)
	assert.Equal(t, successOut, false)
	assert.Equal(t, "Password is either less than the minimum of 8 or over the maximum number of 20 characters", err.Error())
}

func TestLoginFailOverMaxPassword(t *testing.T) {
	client, err := NewAuthClient(url)
	badPasswordIn := "123asdasdasdasdasdasdasdasdadasdasdadaqweqwe4"
	tokenOut, successOut, err := client.Login(baduser, badPasswordIn)

	assert.Len(t, tokenOut, 0)
	assert.Equal(t, successOut, false)
	assert.Equal(t, "Password is either less than the minimum of 8 or over the maximum number of 20 characters", err.Error())
}

func TestLoginFailUserIsNotAlphabetical(t *testing.T) {
	client, err := NewAuthClient(url)
	badUserIn := "1234"
	tokenOut, successOut, err := client.Login(badUserIn, badpassword)

	assert.Len(t, tokenOut, 0)
	assert.Equal(t, successOut, false)
	assert.Equal(t, "Usernames must be alphabetical", err.Error())
}

// IsAuthenticated

func TestIsAuthenticatedFailBadToken(t *testing.T) {
	client, err := NewAuthClient(url)
	usernameIn := baduser
	tokenIn := "shouldntwork"
	success, err := client.IsAuthenticated(usernameIn, tokenIn)

	assert.Equal(t, success, false)
	assert.Equal(t, "Invalid Status Code: 401 Unauthorized", err.Error())
}

// Logout

func TestLogoutFailBadToken(t *testing.T) {
	client, err := NewAuthClient(url)
	usernameIn := baduser
	tokenIn := "shouldntwork"
	success, err := client.Logout(usernameIn, tokenIn)

	assert.Equal(t, success, false)
	assert.Equal(t, "Invalid Status Code: 401 Unauthorized", err.Error())
}

// Happy Path

func TestLogin(t *testing.T) {
	client, err := NewAuthClient(url)
	tokenOut, successOut, err := client.Login(username, password)

	if successOut == true {
		token = tokenOut
	}

	assert.Nil(t, err)
	assert.Equal(t, successOut, true)
	assert.NotNil(t, tokenOut)
}

func TestIsAuthenticated(t *testing.T) {
	client, err := NewAuthClient(url)
	success, err := client.IsAuthenticated(username, token)

	assert.Equal(t, success, true)
	assert.Nil(t, err)
}

func TestLogout(t *testing.T) {
	client, err := NewAuthClient(url)
	success, err := client.Logout(username, token)

	assert.Equal(t, success, true)
	assert.Nil(t, err)
}
