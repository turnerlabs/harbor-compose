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
	shipmentJSON = strings.Replace(shipmentJSON, "${healthcheckTimeout}", "1", 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${healthcheckInterval}", "10", 1)
	t.Log(shipmentJSON)

	//deserialize shipit json
	var shipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipment)
	if err != nil {
		log.Fatal(err)
	}

	//convert shipit model to harbor-compose
	harborCompose := transformShipmentToHarborCompose(&shipment)

	//convert shipit model to docker-compose
	dockerCompose, _ := transformShipmentToDockerCompose(&shipment)

	//debug
	data, _ := yaml.Marshal(dockerCompose)
	t.Log(string(data))

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
	artifacts, err := provider.ProvideArtifacts(&dockerCompose, &harborCompose, "token", "harbor")
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

func TestBuildProviderCircleCIv2_NotSupported(t *testing.T) {
	dockerCompose := DockerCompose{}
	harborCompose := HarborCompose{}

	//load circleciv2 build provider
	provider, err := getBuildProvider("circleciv2")
	if err != nil {
		t.Fail()
	}

	//run the build provider
	_, err = provider.ProvideArtifacts(&dockerCompose, &harborCompose, "token", "notsupported")
	assert.NotNil(t, err, "build provider %s should not be supported")
}

//tests generate --build-provider circleciv2
func TestBuildProviderCircleCIv2_EcsFargate(t *testing.T) {
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
	shipmentJSON = strings.Replace(shipmentJSON, "${healthcheckTimeout}", "1", 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${healthcheckInterval}", "10", 1)
	t.Log(shipmentJSON)

	//deserialize shipit json
	var shipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipment)
	if err != nil {
		log.Fatal(err)
	}

	//convert shipit model to harbor-compose
	harborCompose := transformShipmentToHarborCompose(&shipment)

	//convert shipit model to docker-compose
	dockerCompose, _ := transformShipmentToDockerCompose(&shipment)

	//debug
	data, _ := yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//debug
	data, _ = yaml.Marshal(harborCompose)
	t.Log(string(data))

	svc := dockerCompose.Services[container]
	assert.NotNil(t, svc)

	//load circleciv2 build provider
	provider, err := getBuildProvider("circleciv2")
	if err != nil {
		t.Fail()
	}

	//run the build provider
	artifacts, err := provider.ProvideArtifacts(&dockerCompose, &harborCompose, "token", "ecsfargate")
	if err != nil {
		t.Fail()
	}

	//debug
	data, _ = yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//docker compose configuration should have the build directive
	assert.Equal(t, "../", svc.Build)

	//docker compose configuration should have the circle ci build number in the image tag
	assert.True(t, strings.HasSuffix(svc.Image, "-${CIRCLE_BUILD_NUM}"))

	//docker compose configuration shouldn't have any environment variables
	assert.Equal(t, 0, len(svc.Environment))

	//docker compose configuration shouldn't have any ports
	assert.Equal(t, 0, len(svc.Ports))

	//the provider should output a .circle/config.yml
	assert.NotNil(t, artifacts)
	assert.True(t, len(artifacts) == 2, "expecting 2 artifacts")

	artifact := findBuildArtifact("config.yml", artifacts)
	assert.NotNil(t, artifact)
	t.Log(artifact.FileContents)

	artifact = findBuildArtifact("config.env", artifacts)
	assert.NotNil(t, artifact)
	t.Log(artifact.FileContents)
}

func findBuildArtifact(name string, artifacts []*BuildArtifact) *BuildArtifact {
	for _, b := range artifacts {
		if strings.Contains(b.FilePath, name) {
			return b
		}
	}
	return nil
}

func TestParseDockerImage(t *testing.T) {
	image := "12345678912.dkr.ecr.us-east-1.amazonaws.com/my-service:0.1.0"
	repo, version, pre := parseDockerImage(image)
	t.Log("image = ", image)
	t.Log("repo = ", repo)
	t.Log("version = ", version)
	t.Log("pre = ", pre)

	assert.Equal(t, "12345678912.dkr.ecr.us-east-1.amazonaws.com/my-service", repo)
	assert.Equal(t, "0.1.0", version)
	assert.Equal(t, "", pre)
}

func TestParseDockerImage_PreRelease(t *testing.T) {
	image := "12345678912.dkr.ecr.us-east-1.amazonaws.com/my-service:0.1.0-pre-develop.42"
	repo, version, pre := parseDockerImage(image)
	t.Log("image = ", image)
	t.Log("repo = ", repo)
	t.Log("version = ", version)
	t.Log("pre = ", pre)

	assert.Equal(t, "12345678912.dkr.ecr.us-east-1.amazonaws.com/my-service", repo)
	assert.Equal(t, "0.1.0", version)
	assert.Equal(t, "pre-develop.42", pre)
}

func TestParseDockerImage_Latest(t *testing.T) {
	image := "12345678912.dkr.ecr.us-east-1.amazonaws.com/my-service"
	repo, version, pre := parseDockerImage(image)
	t.Log("image = ", image)
	t.Log("repo = ", repo)
	t.Log("version = ", version)
	t.Log("pre = ", pre)

	assert.Equal(t, "12345678912.dkr.ecr.us-east-1.amazonaws.com/my-service", repo)
	assert.Equal(t, "latest", version)
	assert.Equal(t, "", pre)
}