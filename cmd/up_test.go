package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
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

func TestTransformComposeToShipmentEnvironment(t *testing.T) {

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
	newShipment := transformComposeToShipmentEnvironment(shipmentName, composeShipment, dockerCompose)

	serviceConfig, success := dockerCompose.GetServiceConfig(composeServiceName)
	if !success {
		log.Fatal("error getting service config")
	}

	//lookup transformed shipment container
	shipmentContainer := newShipment.Containers[0]

	//assertions

	//shipment transformations
	assert.Equal(t, shipmentName, newShipment.ParentShipment.Name)
	assert.Equal(t, composeShipment.Env, newShipment.Name)
	assert.Equal(t, composeShipment.Group, newShipment.ParentShipment.Group)
	assert.Equal(t, 0, len(newShipment.EnvVars))

	//container transformations
	assert.Equal(t, composeServiceName, shipmentContainer.Name)
	assert.Equal(t, serviceConfig.Image, shipmentContainer.Image)
	assert.Equal(t, len(serviceConfig.Ports), len(shipmentContainer.Ports))
	assert.Equal(t, serviceConfig.Environment.ToMap()["HEALTHCHECK"], shipmentContainer.Ports[0].Healthcheck)

	//all environment variables specified in docker-compose should get tranformed to shipment container vars
	assert.True(t, assertEnvVarsMatch(t, serviceConfig.Environment.ToMap(), shipmentContainer.EnvVars))
}

func TestTransformComposeToShipmentEnvironmentEnvFile(t *testing.T) {
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
	newShipment := transformComposeToShipmentEnvironment(shipmentName, composeShipment, dockerCompose)

	//lookup container service
	serviceConfig, success := dockerCompose.GetServiceConfig(composeServiceName)
	if !success {
		log.Fatal("error getting service config")
	}

	//lookup transformed shipment container
	shipmentContainer := newShipment.Containers[0]

	//assertions

	//shipment transformations
	assert.Equal(t, shipmentName, newShipment.ParentShipment.Name)
	assert.Equal(t, composeShipment.Env, newShipment.Name)
	assert.Equal(t, composeShipment.Group, newShipment.ParentShipment.Group)
	assert.Equal(t, 0, len(newShipment.EnvVars))

	//container transformations
	assert.Equal(t, composeServiceName, shipmentContainer.Name)
	assert.Equal(t, serviceConfig.Image, shipmentContainer.Image)
	assert.Equal(t, len(serviceConfig.Ports), len(shipmentContainer.Ports))
	assert.Equal(t, serviceConfig.Environment.ToMap()["HEALTHCHECK"], shipmentContainer.Ports[0].Healthcheck)

	//all environment variables specified in docker-compose should get tranformed to shipment container vars
	assert.True(t, assertEnvVarsMatch(t, serviceConfig.Environment.ToMap(), shipmentContainer.EnvVars))

	//clean up
	err = os.Remove(envFileName)
	if err != nil {
		t.Fail()
	}
}

func TestTransformComposeToShipmentEnvironmentDotEnv(t *testing.T) {
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
	newShipment := transformComposeToShipmentEnvironment(shipmentName, composeShipment, dockerCompose)

	//lookup container service
	serviceConfig, success := dockerCompose.GetServiceConfig(composeServiceName)
	if !success {
		log.Fatal("error getting service config")
	}

	//lookup transformed shipment container
	shipmentContainer := newShipment.Containers[0]

	//assertions

	//shipment transformations
	assert.Equal(t, shipmentName, newShipment.ParentShipment.Name)
	assert.Equal(t, composeShipment.Env, newShipment.Name)
	assert.Equal(t, composeShipment.Group, newShipment.ParentShipment.Group)
	assert.Equal(t, 0, len(newShipment.EnvVars))

	//container transformations
	assert.Equal(t, composeServiceName, shipmentContainer.Name)
	assert.Equal(t, serviceConfig.Image, shipmentContainer.Image)
	assert.Equal(t, len(serviceConfig.Ports), len(shipmentContainer.Ports))
	assert.Equal(t, serviceConfig.Environment.ToMap()["HEALTHCHECK"], shipmentContainer.Ports[0].Healthcheck)

	//all environment variables specified in docker-compose should get tranformed to shipment container vars
	assert.True(t, assertEnvVarsMatch(t, serviceConfig.Environment.ToMap(), shipmentContainer.EnvVars))

	//clean up
	err = os.Remove(".env")
	if err != nil {
		t.Fail()
	}
}

func TestTransformComposeToShipmentEnvironmentEnvFileHealthCheck(t *testing.T) {
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
	newShipment := transformComposeToShipmentEnvironment(shipmentName, composeShipment, dockerCompose)

	//lookup container service
	serviceConfig, success := dockerCompose.GetServiceConfig(composeServiceName)
	if !success {
		log.Fatal("error getting service config")
	}

	//lookup transformed shipment container
	shipmentContainer := newShipment.Containers[0]

	//assertions

	//shipment transformations
	assert.Equal(t, shipmentName, newShipment.ParentShipment.Name)
	assert.Equal(t, composeShipment.Env, newShipment.Name)
	assert.Equal(t, composeShipment.Group, newShipment.ParentShipment.Group)
	assert.Equal(t, 0, len(newShipment.EnvVars))

	//container transformations
	assert.Equal(t, composeServiceName, shipmentContainer.Name)
	assert.Equal(t, serviceConfig.Image, shipmentContainer.Image)
	assert.Equal(t, len(serviceConfig.Ports), len(shipmentContainer.Ports))
	assert.Equal(t, serviceConfig.Environment.ToMap()["HEALTHCHECK"], shipmentContainer.Ports[0].Healthcheck)

	//all environment variables specified in docker-compose should get tranformed to shipment container vars
	assert.True(t, assertEnvVarsMatch(t, serviceConfig.Environment.ToMap(), shipmentContainer.EnvVars))

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

//tests new shipment environment name validation
func TestUpValidateEnvName(t *testing.T) {

	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:3000
    environment:
      HEALTHCHECK: /health
      FOO: bar`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev_1
    barge: sandbox
    containers:
    - app
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product`

	//parse the compose yaml into objects that we can work with
	shipmentName := "mss-test-shipment"
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", shipmentName, 1)
	composeServiceName := "app"
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", composeServiceName, 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[shipmentName]
	t.Log(dockerComposeYaml)

	//get a new shipment
	newShipment := transformComposeToShipmentEnvironment(shipmentName, composeShipment, dockerCompose)

	//test func
	err := validateUp(&newShipment, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), messageEnvironmentUnderscores)
}

//tests new shipment container count validation
func TestUpValidateContainerCount(t *testing.T) {

	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:3000
    environment:
      HEALTHCHECK: /health
      FOO: bar`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev
    barge: sandbox
    containers:
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product`

	//parse the compose yaml into objects that we can work with
	shipmentName := "mss-test-shipment"
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", shipmentName, 1)
	composeServiceName := "app"
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", composeServiceName, 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[shipmentName]
	t.Log(dockerComposeYaml)

	//get a new shipment
	newShipment := transformComposeToShipmentEnvironment(shipmentName, composeShipment, dockerCompose)

	//test func
	err := validateUp(&newShipment, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), messageContainerRequired)
}

//tests up (happy path)
func TestUpValidate(t *testing.T) {

	//create an existing shipment
	shipmentJSON := getSampleShipmentJSONForValidation()

	//update json with test values
	name := "mss-poc-app"
	env := "dev"
	barge := "digital-sandbox"
	replicas := 2
	container := "web"

	shipmentJSON = strings.Replace(shipmentJSON, "${name}", name, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${env}", env, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${barge}", barge, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${replicas}", strconv.Itoa(replicas), 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${container}", container, 1)
	t.Log(shipmentJSON)

	//deserialize shipit json
	var existingShipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &existingShipment)
	check(err)

	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:${port}
    environment:
      HEALTHCHECK: ${healthcheck}
      FOO: bar`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev
    barge: ${barge}
    containers:
    - ${composeServiceName}
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product`

	//parse the compose yaml into objects that we can work with
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", name, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${composeServiceName}", container, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${barge}", barge, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", container, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${port}", "5000", 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${healthcheck}", "/hc", 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[name]
	t.Log(dockerComposeYaml)
	t.Log(harborComposeYaml)

	//make changes to shipment
	desiredShipment := transformComposeToShipmentEnvironment(name, composeShipment, dockerCompose)

	//test func
	err = validateUp(&desiredShipment, &existingShipment)

	//no errors
	assert.Nil(t, err)
}

//tests up with container name change
func TestUpValidateContainerNameChange(t *testing.T) {

	//create an existing shipment
	shipmentJSON := getSampleShipmentJSONForValidation()

	//update json with test values
	name := "mss-poc-app"
	env := "dev"
	barge := "digital-sandbox"
	replicas := 2
	container := "web"

	shipmentJSON = strings.Replace(shipmentJSON, "${name}", name, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${env}", env, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${barge}", barge, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${replicas}", strconv.Itoa(replicas), 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${container}", container, 1)
	t.Log(shipmentJSON)

	//deserialize shipit json
	var existingShipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &existingShipment)
	check(err)

	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:${port}
    environment:
      HEALTHCHECK: ${healthcheck}
      FOO: bar`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev
    barge: ${barge}
    containers:
    - ${composeServiceName}
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product`

	//simulate changing container name
	container = "change"

	//parse the compose yaml into objects that we can work with
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", name, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${composeServiceName}", container, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${barge}", barge, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", container, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${port}", "5000", 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${healthcheck}", "/hc", 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[name]
	t.Log(dockerComposeYaml)
	t.Log(harborComposeYaml)

	//make changes to shipment
	desiredShipment := transformComposeToShipmentEnvironment(name, composeShipment, dockerCompose)

	//test func
	err = validateUp(&desiredShipment, &existingShipment)

	//look for error
	t.Log(err)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), messageChangeContainer)
}

//tests up with port change
func TestUpValidatePortChange(t *testing.T) {

	//create an existing shipment
	shipmentJSON := getSampleShipmentJSONForValidation()

	//update json with test values
	name := "mss-poc-app"
	env := "dev"
	barge := "digital-sandbox"
	replicas := 2
	container := "web"

	shipmentJSON = strings.Replace(shipmentJSON, "${name}", name, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${env}", env, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${barge}", barge, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${replicas}", strconv.Itoa(replicas), 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${container}", container, 1)
	t.Log(shipmentJSON)

	//deserialize shipit json
	var existingShipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &existingShipment)
	check(err)

	dockerComposeYaml := `
  version: "2"
  services:
    ${composeServiceName}:
      image: registry/app:1.0
      ports:
      - 80:${port}
      environment:
        HEALTHCHECK: ${healthcheck}
        FOO: bar`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev
    barge: ${barge}
    containers:
    - ${composeServiceName}
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product`

	//parse the compose yaml into objects that we can work with
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", name, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${composeServiceName}", container, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${barge}", barge, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", container, 1)

	//change port
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${port}", "8000", 1)

	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${healthcheck}", "/hc", 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[name]
	t.Log(dockerComposeYaml)
	t.Log(harborComposeYaml)

	//make changes to shipment
	desiredShipment := transformComposeToShipmentEnvironment(name, composeShipment, dockerCompose)

	//test func
	err = validateUp(&desiredShipment, &existingShipment)

	//look for error
	t.Log(err)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), messageChangePort)
}

//tests up with adding a port
func TestUpValidatePortAdd(t *testing.T) {

	//create an existing shipment
	shipmentJSON := getSampleShipmentJSONForValidation()

	//update json with test values
	name := "mss-poc-app"
	env := "dev"
	barge := "digital-sandbox"
	replicas := 2
	container := "web"

	shipmentJSON = strings.Replace(shipmentJSON, "${name}", name, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${env}", env, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${barge}", barge, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${replicas}", strconv.Itoa(replicas), 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${container}", container, 1)
	t.Log(shipmentJSON)

	//deserialize shipit json
	var existingShipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &existingShipment)
	check(err)

	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:${port}
    - 81:3000
    environment:
      HEALTHCHECK: ${healthcheck}
      FOO: bar`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev
    barge: ${barge}
    containers:
    - ${composeServiceName}
    replicas: 2
    group: mss
    property: turner
    project: project
    product: product`

	//parse the compose yaml into objects that we can work with
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", name, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${composeServiceName}", container, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${barge}", barge, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", container, 1)

	//change port
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${port}", "8000", 1)

	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${healthcheck}", "/hc", 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[name]
	t.Log(dockerComposeYaml)
	t.Log(harborComposeYaml)

	//make changes to shipment
	desiredShipment := transformComposeToShipmentEnvironment(name, composeShipment, dockerCompose)

	//test func
	err = validateUp(&desiredShipment, &existingShipment)

	//look for error
	t.Log(err)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), messageChangePort)
}

//tests up with health check change
func TestUpValidateHealthCheckChange(t *testing.T) {

	//create an existing shipment
	shipmentJSON := getSampleShipmentJSONForValidation()

	//update json with test values
	name := "mss-poc-app"
	env := "dev"
	barge := "digital-sandbox"
	replicas := 2
	container := "web"

	shipmentJSON = strings.Replace(shipmentJSON, "${name}", name, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${env}", env, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${barge}", barge, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${replicas}", strconv.Itoa(replicas), 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${container}", container, 1)
	t.Log(shipmentJSON)

	//deserialize shipit json
	var existingShipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &existingShipment)
	check(err)

	dockerComposeYaml := `
  version: "2"
  services:
    ${composeServiceName}:
      image: registry/app:1.0
      ports:
      - 80:${port}
      environment:
        HEALTHCHECK: ${healthcheck}
        FOO: bar`

	harborComposeYaml := `
  shipments:
    ${shipmentName}:
      env: dev
      barge: ${barge}
      containers:
      - ${composeServiceName}
      replicas: 2
      group: mss
      property: turner
      project: project
      product: product`

	//parse the compose yaml into objects that we can work with
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", name, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${composeServiceName}", container, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${barge}", barge, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", container, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${port}", "5000", 1)

	//change health check
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${healthcheck}", "change", 1)

	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[name]
	t.Log(dockerComposeYaml)
	t.Log(harborComposeYaml)

	//make changes to shipment
	desiredShipment := transformComposeToShipmentEnvironment(name, composeShipment, dockerCompose)

	//test func
	err = validateUp(&desiredShipment, &existingShipment)

	//look for error
	t.Log(err)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), messageChangeHealthCheck)
}

//tests up with health check interval == timeout
func TestUpValidateHealthCheckIntervalEqualToTimeout(t *testing.T) {

	//update json with test values
	name := "mss-poc-app"
	container := "web"

	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:5000
    environment:
      HEALTHCHECK: hc
      FOO: bar`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev
    barge: digital-sandbox
    containers:
    - ${composeServiceName}
    replicas: 2
    group: mss
    healthcheckIntervalSeconds: 10
    healthcheckTimeoutSeconds: 10`

	//parse the compose yaml into objects that we can work with
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", name, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${composeServiceName}", container, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", container, 1)

	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[name]
	t.Log(dockerComposeYaml)
	t.Log(harborComposeYaml)

	//test validate
	desiredShipment := transformComposeToShipmentEnvironment(name, composeShipment, dockerCompose)
	err := validateUp(&desiredShipment, nil)

	//expect to fail (interval == timeout)
	t.Log(err)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), messageIntervalGreaterThanTimeout)
}

//tests up with health check interval > timeout
func TestUpValidateHealthCheckIntervalGreaterThanTimeout(t *testing.T) {

	//update json with test values
	name := "mss-poc-app"
	container := "web"

	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:5000
    environment:
      HEALTHCHECK: hc
      FOO: bar`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev
    barge: digital-sandbox
    containers:
    - ${composeServiceName}
    replicas: 2
    group: mss
    healthcheckIntervalSeconds: 11
    healthcheckTimeoutSeconds: 10`

	//parse the compose yaml into objects that we can work with
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", name, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${composeServiceName}", container, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", container, 1)

	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[name]
	t.Log(dockerComposeYaml)
	t.Log(harborComposeYaml)

	//test validate
	desiredShipment := transformComposeToShipmentEnvironment(name, composeShipment, dockerCompose)
	err := validateUp(&desiredShipment, nil)

	//expect to pass (interval > timeout)
	t.Log(err)
	assert.Nil(t, err)
}

//tests up with health check interval < timeout
func TestUpValidateHealthCheckIntervalLessThanTimeout(t *testing.T) {

	//update json with test values
	name := "mss-poc-app"
	container := "web"

	dockerComposeYaml := `
version: "2"
services:
  ${composeServiceName}:
    image: registry/app:1.0
    ports:
    - 80:5000
    environment:
      HEALTHCHECK: hc
      FOO: bar`

	harborComposeYaml := `
shipments:
  ${shipmentName}:
    env: dev
    barge: digital-sandbox
    containers:
      - ${composeServiceName}
    replicas: 2
    group: mss
    healthcheckIntervalSeconds: 9
    healthcheckTimeoutSeconds: 10`

	//parse the compose yaml into objects that we can work with
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", name, 1)
	harborComposeYaml = strings.Replace(harborComposeYaml, "${composeServiceName}", container, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", container, 1)

	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[name]
	t.Log(dockerComposeYaml)
	t.Log(harborComposeYaml)

	//test validate
	desiredShipment := transformComposeToShipmentEnvironment(name, composeShipment, dockerCompose)
	err := validateUp(&desiredShipment, nil)

	//expect to fail (interval < timeout)
	t.Log(err)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), messageIntervalGreaterThanTimeout)
}

func TestTransformComposeToShipmentEnvironmentHiddenEnvFile(t *testing.T) {
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
    env_file: 
    - ${envFileName}
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

	//write hidden.env file containing env vars
	envFileName := "/tmp/hidden.env"
	err := ioutil.WriteFile(envFileName, []byte("HIDDEN=foo"), 0644)
	if err != nil {
		t.Fail()
	}

	shipmentName := "mss-test-shipment"
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", shipmentName, 1)
	composeServiceName := "app"
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", composeServiceName, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${envFileName}", envFileName, 1)
	dockerCompose, _ := unmarshalCompose(dockerComposeYaml, harborComposeYaml)

	//lookup container service
	serviceConfig, success := dockerCompose.GetServiceConfig(composeServiceName)
	if !success {
		log.Fatal("error getting service config")
	}

	harborEnvVars := transformDockerServiceEnvVarsToHarborEnvVars(serviceConfig)

	//assertions

	//all environment variables specified in docker-compose should get tranformed to shipment container vars
	assert.True(t, assertEnvVarsMatch(t, serviceConfig.Environment.ToMap(), harborEnvVars))

	//vars in hidden.env should be set to type=hidden
	hiddenVar := getEnvVar("HIDDEN", harborEnvVars)
	assert.Equal(t, "hidden", hiddenVar.Type)

	//clean up
	err = os.Remove(envFileName)
	if err != nil {
		t.Fail()
	}
}

func TestTransformComposeToShipmentEnvironmentHiddenEnvFileWithAnotherEnvFile(t *testing.T) {
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
    env_file: 
    - ${envFileName}
    - ${envFileName2}
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

	//write hidden.env file containing env vars
	envFileName := "/tmp/hidden.env"
	err := ioutil.WriteFile(envFileName, []byte("HIDDEN=foo"), 0644)
	if err != nil {
		t.Fail()
	}

	//also, write another env_file
	envFileName2 := "/tmp/vars.env"
	err = ioutil.WriteFile(envFileName2, []byte("FOOBAR=foo"), 0644)
	if err != nil {
		t.Fail()
	}

	shipmentName := "mss-test-shipment"
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", shipmentName, 1)
	composeServiceName := "app"
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", composeServiceName, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${envFileName}", envFileName, 1)
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${envFileName2}", envFileName2, 1)
	dockerCompose, _ := unmarshalCompose(dockerComposeYaml, harborComposeYaml)

	//lookup container service
	serviceConfig, success := dockerCompose.GetServiceConfig(composeServiceName)
	if !success {
		log.Fatal("error getting service config")
	}

	//test
	harborEnvVars := transformDockerServiceEnvVarsToHarborEnvVars(serviceConfig)

	//assertions

	//all environment variables specified in docker-compose should get tranformed to shipment container vars
	assert.True(t, assertEnvVarsMatch(t, serviceConfig.Environment.ToMap(), harborEnvVars))

	//vars in hidden.env should be set to type=hidden
	hiddenVar := getEnvVar("HIDDEN", harborEnvVars)
	assert.Equal(t, "hidden", hiddenVar.Type)

	anotherVar := getEnvVar("FOOBAR", harborEnvVars)
	assert.Equal(t, "basic", anotherVar.Type)

	//clean up
	err = os.Remove(envFileName)
	if err != nil {
		t.Fail()
	}
	err = os.Remove(envFileName2)
	if err != nil {
		t.Fail()
	}
}

//ensures up correctly translates enableMonitoring
func TestUpEnableMonitoringDefault(t *testing.T) {

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
	newShipment := transformComposeToShipmentEnvironment(shipmentName, composeShipment, dockerCompose)

	//enableMonitoring should default to true if not specified in yaml
	assert.True(t, newShipment.EnableMonitoring, "expecting enableMonitoring to default to true")
}

//ensures up correctly translates enableMonitoring
func TestUpEnableMonitoring(t *testing.T) {

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
    enableMonitoring: false
`

	//parse the compose yaml into objects that we can work with
	shipmentName := "mss-test-shipment"
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", shipmentName, 1)
	composeServiceName := "app"
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", composeServiceName, 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[shipmentName]

	//test func
	newShipment := transformComposeToShipmentEnvironment(shipmentName, composeShipment, dockerCompose)

	//enableMonitoring should get mapped from yaml
	assert.False(t, newShipment.EnableMonitoring, "expecting enableMonitoring to be false")
}

func TestUpHealthcheckSettings(t *testing.T) {

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
    enableMonitoring: true
    healthcheckTimeoutSeconds: 10
    healthcheckIntervalSeconds: 100
`

	//parse the compose yaml into objects that we can work with
	shipmentName := "mss-test-shipment"
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", shipmentName, 1)
	composeServiceName := "app"
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", composeServiceName, 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[shipmentName]

	fmt.Println(harborCompose.Shipments[shipmentName].HealthcheckTimeoutSeconds)

	//test func
	newShipment := transformComposeToShipmentEnvironment(shipmentName, composeShipment, dockerCompose)

	//lookup transformed shipment container
	shipmentContainer := newShipment.Containers[0]

	//assertions
	assert.Equal(t, 10, *shipmentContainer.Ports[0].HealthcheckTimeout)
	assert.Equal(t, 100, *shipmentContainer.Ports[0].HealthcheckInterval)
}

//test interval > timeout
func TestUpHealthcheckSettingsInterval(t *testing.T) {

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
    enableMonitoring: true
    healthcheckTimeoutSeconds: 10
    healthcheckIntervalSeconds: 9
`

	//parse the compose yaml into objects that we can work with
	shipmentName := "mss-test-shipment"
	harborComposeYaml = strings.Replace(harborComposeYaml, "${shipmentName}", shipmentName, 1)
	composeServiceName := "app"
	dockerComposeYaml = strings.Replace(dockerComposeYaml, "${composeServiceName}", composeServiceName, 1)
	dockerCompose, harborCompose := unmarshalCompose(dockerComposeYaml, harborComposeYaml)
	composeShipment := harborCompose.Shipments[shipmentName]

	fmt.Println(harborCompose.Shipments[shipmentName].HealthcheckTimeoutSeconds)

	//test func
	newShipment := transformComposeToShipmentEnvironment(shipmentName, composeShipment, dockerCompose)
	err := validateUp(&newShipment, nil)

	t.Log(err)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), messageIntervalGreaterThanTimeout)
}

func TestEmptyEnvVar(t *testing.T) {

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
    env_file: 
    - ${envFileName}
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

	//write hidden.env file containing empty env vars
	envFileName := "/tmp/hidden.env"
	err := ioutil.WriteFile(envFileName, []byte("HIDDEN="), 0644)
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

	//test validate
	desiredShipment := transformComposeToShipmentEnvironment(shipmentName, composeShipment, dockerCompose)
	err = validateUp(&desiredShipment, nil)

	//expect to fail since env var is empty
	t.Log(err)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), messageEnvvarsCannotBeEmpty)

	//clean up
	err = os.Remove(envFileName)
	if err != nil {
		t.Fail()
	}
}
