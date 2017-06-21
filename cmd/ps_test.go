package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetShipmentEndpointHttp(t *testing.T) {
	endpoint := getShipmentEndpoint("my-shipment", "dev", "ec2", "80")
	t.Log(endpoint)
	assert.Equal(t, "http://my-shipment.dev.services.ec2.dmtio.net", endpoint)
}

func TestGetShipmentEndpointHttps(t *testing.T) {
	endpoint := getShipmentEndpoint("my-shipment", "dev", "ec2", "443")
	t.Log(endpoint)
	assert.Equal(t, "https://my-shipment.dev.services.ec2.dmtio.net", endpoint)
}

func TestGetShipmentEndpointNon80(t *testing.T) {
	endpoint := getShipmentEndpoint("my-shipment", "dev", "ec2", "5000")
	t.Log(endpoint)
	assert.Equal(t, "http://my-shipment.dev.services.ec2.dmtio.net:5000", endpoint)
}
