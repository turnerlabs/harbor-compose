package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// GetShipmentEnvironment

var shipment string
var env string

func TestGetShipmentEnvironment(t *testing.T) {
	t.SkipNow()
	var token = ""
	shipment = "ams-harbor-api-api"
	env = "prod"

	if len(token) > 0 {
		shipmentEnv := GetShipmentEnvironment(shipment, env, token)

		assert.Nil(t, shipmentEnv)
		// assert.NotNil(t, shipmentEnv)
		// assert.Equal(t, shipmentEnv.Name, "prod")
	}
}
