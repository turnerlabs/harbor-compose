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

func TestGetShipmentPrimaryPort(t *testing.T) {

	dockerComposeYaml := `
version: "2"
services:
  app:
    image: registry/app:1.0
    ports:
      - 80:3000
    environment:
      HEALTHCHECK: /health
`

	harborComposeYaml := `
shipments:
  app:
    env: dev
    barge: sandbox
    containers:
      - app 
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product	
`

	//parse the compose yaml into objects that we can work with
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)

	//test that correct port is returned
	for _, shipment := range harborCompose.Shipments {
		port, err := getShipmentPrimaryPort(dockerCompose, shipment)
		assert.Equal(t, "80", port)
		assert.Nil(t, err)
	}
}

func TestGetShipmentPrimaryPortAltFormat(t *testing.T) {

	dockerComposeYaml := `
version: "2"
services:
  app:
    image: registry/app:1.0
    ports:
      - "443:3000"
    environment:
      HEALTHCHECK: /health
`

	harborComposeYaml := `
shipments:
  app:
    env: dev
    barge: sandbox
    containers:
      - app 
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product	
`

	//parse the compose yaml into objects that we can work with
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)

	//test that correct port is returned
	for _, shipment := range harborCompose.Shipments {
		port, err := getShipmentPrimaryPort(dockerCompose, shipment)
		assert.Equal(t, "443", port)
		assert.Nil(t, err)
	}
}

func TestGetShipmentPrimaryPortNoPorts(t *testing.T) {
	dockerComposeYaml := `
version: "2"
services:
  app:
    image: registry/app:1.0
    environment:
      HEALTHCHECK: /health
`

	harborComposeYaml := `
shipments:
  app:
    env: dev
    barge: sandbox
    containers:
      - app 
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product	
`

	//parse the compose yaml into objects that we can work with
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)

	//test that we get an error
	for _, shipment := range harborCompose.Shipments {
		_, err := getShipmentPrimaryPort(dockerCompose, shipment)
		assert.NotNil(t, err)
	}
}

func TestGetShipmentPrimaryPortNoService(t *testing.T) {
	dockerComposeYaml := `
version: "2"
services:
  app:
    image: registry/app:1.0
    ports:
      - 80:3000		
    environment:
      HEALTHCHECK: /health
`

	harborComposeYaml := `
shipments:
  app:
    env: dev
    barge: sandbox
    containers:
      - invalid-app
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product	
`

	//parse the compose yaml into objects that we can work with
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)

	//test that we get an error
	for _, shipment := range harborCompose.Shipments {
		_, err := getShipmentPrimaryPort(dockerCompose, shipment)
		assert.NotNil(t, err)
	}
}
