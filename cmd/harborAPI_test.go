package cmd

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turnerlabs/harbor-auth-client"
)

var token string

var url string
var username string
var password string

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

	// Since all functions need auth, this may be the best implmentation
	client, err := harborauth.NewAuthClient(url)
	tokenIn, successOut, err := client.Login(username, password)
	if err != nil && successOut != true {
		fmt.Println("Unable to Login.")
		os.Exit(0)
	} else {
		token = tokenIn
	}

	os.Exit(m.Run())
}

// GetShipmentEnvironment

var shipment string
var env string

func TestGetShipmentEnvironment(t *testing.T) {
	shipment = "ams-harbor-api-api"
	env = "prod"

	if len(token) > 0 {
		shipmentEnv := GetShipmentEnvironment(shipment, env, token)

		assert.NotNil(t, shipmentEnv)
		assert.Equal(t, shipmentEnv.Name, "prod")
	}
}
