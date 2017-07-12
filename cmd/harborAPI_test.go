package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetShipmentEnvironment(t *testing.T) {
	if !*integrationTest || *usernameTest == "" || *passwordTest == "" {
		t.SkipNow()
	}

	os.Unsetenv("HC_CONFIG")

	//login
	token, err := harborLogin(*usernameTest, *passwordTest)
	assert.NotEmpty(t, token)
	assert.Nil(t, err)

	//todo: consider creating a shipment here

	//test
	shipment := "mss-shipit-api"
	env := "dev"
	shipmentEnv := GetShipmentEnvironment(*usernameTest, token, shipment, env)

	//assertions
	assert.NotNil(t, shipmentEnv)
	assert.Equal(t, shipmentEnv.ParentShipment.Name, shipment)
	assert.Equal(t, shipmentEnv.Name, env)

	//logout
	success, err := harborLogout(*usernameTest, token)
	assert.Nil(t, err)
	assert.True(t, success)
}
