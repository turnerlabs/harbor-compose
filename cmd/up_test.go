package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpEnvVarEqualsSign(t *testing.T) {

	yaml := `
version: "2"
services:
  container:
    image: helloworld
    environment:
      CHAR_EQUAL: foo=bar	
`

	dockerCompose := unmarshalDockerCompose(yaml)
	proj, _ := dockerCompose.GetServiceConfig("container")

	assert.Equal(t, "foo=bar", proj.Environment.ToMap()["CHAR_EQUAL"])
}

func TestTransformComposeToNewShipment(t *testing.T) {

	//test compose yaml transformation to a new harbor shipment
	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:3000
    environment:
      HEALTHCHECK: /health
      FOO: bar
`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
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
	shipmentName := "mss-test-shipment"
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", shipmentName, 1)
	composeServiceName := "app"
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", composeServiceName, 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[shipmentName]

	//test func
	newShipment := transformComposeToNewShipment(shipmentName, composeShipment, dockerCompose)

	serviceConfig, success := dockerCompose.GetServiceConfig(composeServiceName)
	if !success {
		log.Fatal("error getting service config")
	}

	//lookup transformed shipment container
	shipmentContainer := newShipment.Containers[0]

	//assertions

	//shipment transformations
	assert.Equal(t, shipmentName, newShipment.Info.Name)
	assert.Equal(t, composeShipment.Env, newShipment.Environment.Name)
	assert.Equal(t, composeShipment.Group, newShipment.Info.Group)
	assert.Equal(t, 0, len(newShipment.Environment.Vars))

	//container transformations
	assert.Equal(t, composeServiceName, shipmentContainer.Name)
	assert.Equal(t, serviceConfig.Image, shipmentContainer.Image)
	assert.Equal(t, len(serviceConfig.Ports), len(shipmentContainer.Ports))
	assert.Equal(t, serviceConfig.Environment.ToMap()["HEALTHCHECK"], shipmentContainer.Ports[0].Healthcheck)

	//all environment variables specified in docker-compose should get tranformed to shipment container vars
	assert.True(t, assertEnvVarsMatch(t, serviceConfig.Environment.ToMap(), shipmentContainer.Vars))
}

func TestTransformComposeToNewShipmentEnvFile(t *testing.T) {
	//test compose yaml transformation to a new harbor shipment using env_file

	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:3000
    environment:
      HEALTHCHECK: /health
      FOO: bar
    env_file: ${envFileName}
`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev
    barge: corp-sandbox
    containers:
    - app
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product
`

	//write test.env file containing env vars
	envFileName := fmt.Sprintf("/tmp/%v.env", t.Name())
	t.Log(envFileName)
	err := ioutil.WriteFile(envFileName, []byte("FROM_ENV_FILE=foo"), 0644)
	if err != nil {
		t.Fail()
	}

	shipmentName := "mss-test-shipment"
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", shipmentName, 1)
	composeServiceName := "app"
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", composeServiceName, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${envFileName}", envFileName, 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[shipmentName]

	//test func
	newShipment := transformComposeToNewShipment(shipmentName, composeShipment, dockerCompose)

	//lookup container service
	serviceConfig, success := dockerCompose.GetServiceConfig(composeServiceName)
	if !success {
		log.Fatal("error getting service config")
	}

	//lookup transformed shipment container
	shipmentContainer := newShipment.Containers[0]

	//assertions

	//shipment transformations
	assert.Equal(t, shipmentName, newShipment.Info.Name)
	assert.Equal(t, composeShipment.Env, newShipment.Environment.Name)
	assert.Equal(t, composeShipment.Group, newShipment.Info.Group)
	assert.Equal(t, 0, len(newShipment.Environment.Vars))

	//container transformations
	assert.Equal(t, composeServiceName, shipmentContainer.Name)
	assert.Equal(t, serviceConfig.Image, shipmentContainer.Image)
	assert.Equal(t, len(serviceConfig.Ports), len(shipmentContainer.Ports))
	assert.Equal(t, serviceConfig.Environment.ToMap()["HEALTHCHECK"], shipmentContainer.Ports[0].Healthcheck)

	//all environment variables specified in docker-compose should get tranformed to shipment container vars
	assert.True(t, assertEnvVarsMatch(t, serviceConfig.Environment.ToMap(), shipmentContainer.Vars))

	//clean up
	err = os.Remove(envFileName)
	if err != nil {
		t.Fail()
	}
}

func TestTransformComposeToNewShipmentDotEnv(t *testing.T) {
	//test compose yaml transformation to a new harbor shipment using .env file

	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:3000
    environment:
      HEALTHCHECK: /health
      FOO: bar
      FROM_DOT_ENV: ${FROM_DOT_ENV}
`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev
    barge: corp-sandbox
    containers:
    - app
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product
`

	//write .env file containing env vars, to current directory
	err := ioutil.WriteFile(".env", []byte("FROM_DOT_ENV=foo"), 0644)
	if err != nil {
		t.Fail()
	}

	shipmentName := "mss-test-shipment"
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", shipmentName, 1)
	composeServiceName := "app"
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", composeServiceName, 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[shipmentName]

	//test func
	newShipment := transformComposeToNewShipment(shipmentName, composeShipment, dockerCompose)

	//lookup container service
	serviceConfig, success := dockerCompose.GetServiceConfig(composeServiceName)
	if !success {
		log.Fatal("error getting service config")
	}

	//lookup transformed shipment container
	shipmentContainer := newShipment.Containers[0]

	//assertions

	//shipment transformations
	assert.Equal(t, shipmentName, newShipment.Info.Name)
	assert.Equal(t, composeShipment.Env, newShipment.Environment.Name)
	assert.Equal(t, composeShipment.Group, newShipment.Info.Group)
	assert.Equal(t, 0, len(newShipment.Environment.Vars))

	//container transformations
	assert.Equal(t, composeServiceName, shipmentContainer.Name)
	assert.Equal(t, serviceConfig.Image, shipmentContainer.Image)
	assert.Equal(t, len(serviceConfig.Ports), len(shipmentContainer.Ports))
	assert.Equal(t, serviceConfig.Environment.ToMap()["HEALTHCHECK"], shipmentContainer.Ports[0].Healthcheck)

	//all environment variables specified in docker-compose should get tranformed to shipment container vars
	assert.True(t, assertEnvVarsMatch(t, serviceConfig.Environment.ToMap(), shipmentContainer.Vars))

	//clean up
	err = os.Remove(".env")
	if err != nil {
		t.Fail()
	}
}

func TestTransformComposeToNewShipmentEnvFileHealthCheck(t *testing.T) {
	//test use of health check in env_file

	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:3000
    environment:
      FOO: bar
    env_file: ${envFileName}
`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev
    barge: corp-sandbox
    containers:
    - app
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product
`

	//write test.env file containing env vars
	envFileName := fmt.Sprintf("/tmp/%v.env", t.Name())
	t.Log(envFileName)
	err := ioutil.WriteFile(envFileName, []byte("HEALTHCHECK=/health"), 0644)
	if err != nil {
		t.Fail()
	}

	shipmentName := "mss-test-shipment"
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", shipmentName, 1)
	composeServiceName := "app"
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", composeServiceName, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${envFileName}", envFileName, 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[shipmentName]

	//test func
	newShipment := transformComposeToNewShipment(shipmentName, composeShipment, dockerCompose)

	//lookup container service
	serviceConfig, success := dockerCompose.GetServiceConfig(composeServiceName)
	if !success {
		log.Fatal("error getting service config")
	}

	//lookup transformed shipment container
	shipmentContainer := newShipment.Containers[0]

	//assertions

	//shipment transformations
	assert.Equal(t, shipmentName, newShipment.Info.Name)
	assert.Equal(t, composeShipment.Env, newShipment.Environment.Name)
	assert.Equal(t, composeShipment.Group, newShipment.Info.Group)
	assert.Equal(t, 0, len(newShipment.Environment.Vars))

	//container transformations
	assert.Equal(t, composeServiceName, shipmentContainer.Name)
	assert.Equal(t, serviceConfig.Image, shipmentContainer.Image)
	assert.Equal(t, len(serviceConfig.Ports), len(shipmentContainer.Ports))
	assert.Equal(t, serviceConfig.Environment.ToMap()["HEALTHCHECK"], shipmentContainer.Ports[0].Healthcheck)

	//all environment variables specified in docker-compose should get tranformed to shipment container vars
	assert.True(t, assertEnvVarsMatch(t, serviceConfig.Environment.ToMap(), shipmentContainer.Vars))

	//clean up
	err = os.Remove(envFileName)
	if err != nil {
		t.Fail()
	}
}

func assertEnvVarsMatch(t *testing.T, composeEnvVars map[string]string, shipmentEnvVars []EnvVarPayload) bool {
	//look for each composeEnvVar in shipmentEnvVars
	for envVarName, envVarValue := range composeEnvVars {
		//find an object in the slice with matching name
		found := false
		for _, shipmentEnvVar := range shipmentEnvVars {
			if envVarName == shipmentEnvVar.Name && envVarValue == shipmentEnvVar.Value {
				found = true
			}
		}
		if !found {
			t.Logf("env var %v not found in shipment", envVarName)
			return false
		}
	}

	//look for each shipmentEnvVar in composeEnvVars
	for _, v := range shipmentEnvVars {
		if composeEnvVars[v.Name] != v.Value {
			t.Logf("env var %v not found in compose", v.Name)
			return false
		}
	}

	return true
}
