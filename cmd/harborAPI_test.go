package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Flags for testing

func TestMain(m *testing.M) {
	// flag.StringVar(&url, "url", "", "url for authorization")
	// flag.Parse()
	//
	// if len(url) == 0 {
	// 	fmt.Println("Missing url.")
	// 	os.Exit(0)
	// }
	//
	// os.Exit(m.Run())
}

// GetShipmentEnvironment

var shipment string
var env string
var token string

func TestGetShipmentEnvironment(t *testing.T) {
	shipmentEnv := GetShipmentEnvironment(shipment, env, token)

	assert.NotNil(t, shipmentEnv)
	assert.Equal(t, shipmentEnv.Name, "test")
}
