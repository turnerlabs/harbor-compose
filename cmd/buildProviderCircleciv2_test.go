package cmd

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

//tests generate --build-provider circleciv2
func TestBuildProviderCircleCIv2(t *testing.T) {
	shipmentJSON := getSampleShipmentJSON()

	//update json with test values
	name := "mss-poc-app"
	env := "dev"
	barge := "digital-sandbox"
	replicas := 2
	group := "mss"
	foo := "bar"
	project := "project"
	property := "property"
	product := "product"
	envLevel := "ENV_LEVEL"
	containerLevel := "CONTAINER_LEVEL"
	container := "web"

	shipmentJSON = strings.Replace(shipmentJSON, "${name}", name, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${env}", env, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${barge}", barge, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${replicas}", strconv.Itoa(replicas), 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${group}", group, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${foo}", foo, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${property}", property, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${product}", product, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${project}", project, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${envLevel}", envLevel, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${containerLevel}", containerLevel, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${container}", container, 1)
	t.Log(shipmentJSON)

	//deserialize shipit json
	var shipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipment)
	if err != nil {
		log.Fatal(err)
	}

	//first convert shipit model to docker-compose
	dockerCompose := transformShipmentToDockerCompose(&shipment)

	//debug
	data, _ := yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//test
	harborCompose := transformShipmentToHarborCompose(&shipment, &dockerCompose)

	//debug
	data, _ = yaml.Marshal(harborCompose)
	t.Log(string(data))

	svc := dockerCompose.Services[container]
	assert.NotNil(t, svc)

	//load circleciv1 build provider
	provider, err := getBuildProvider("circleciv2")
	if err != nil {
		t.Fail()
	}

	//run the build provider
	artifacts, err := provider.ProvideArtifacts(&dockerCompose, &harborCompose, "token")
	if err != nil {
		t.Fail()
	}

	//debug
	data, _ = yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//docker compose configuration should have the build directive
	assert.NotEmpty(t, svc.Build)

	//docker compose configuration should have the circle ci build number in the image tag
	assert.True(t, strings.HasSuffix(svc.Image, "-${CIRCLE_BUILD_NUM}"))

	//docker compose configuration shouldn't have any environment variables
	assert.Equal(t, 0, len(svc.Environment))

	//docker compose configuration shouldn't have any ports
	assert.Equal(t, 0, len(svc.Ports))

	//the provider should output a .circle/config.yml
	assert.NotNil(t, artifacts)
	assert.Equal(t, ".circleci/config.yml", artifacts[0].FilePath)
	t.Log(artifacts[0].FileContents)
}
