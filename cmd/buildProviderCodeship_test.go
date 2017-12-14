package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

//tests generate --build-provider codeship
func TestBuildProviderCodeship(t *testing.T) {
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
	check(err)

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
	provider, err := getBuildProvider("codeship")
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

	//docker compose configuration should have the codeship build number in the image tag
	assert.True(t, strings.HasSuffix(svc.Image, "-${CI_BUILD_ID}"))

	//docker compose configuration shouldn't have any environment variables
	assert.Equal(t, 0, len(svc.Environment))

	//docker compose configuration shouldn't have any ports
	assert.Equal(t, 0, len(svc.Ports))

	assert.NotNil(t, artifacts)
	assertArtifact(t, artifacts, "codeship-services.yml")
	assertArtifact(t, artifacts, "codeship-steps.yml")
	assertArtifact(t, artifacts, "codeship.env")
	assertArtifact(t, artifacts, "codeship.aes")

	//asert that docker-push.sh is writable
	dockerPush := assertArtifact(t, artifacts, "docker-push.sh")
	assert.True(t, dockerPush.FileMode == 0777, "expecting docker-push.sh file mode to be executable")

	//assert that codeship.env and codeship.aes are added to .gitignore/.dockerignore
	gitignorebits, err := ioutil.ReadFile(".gitignore")
	check(err)
	gitignore := string(gitignorebits)
	assert.Contains(t, gitignore, "codeship.env")
	assert.Contains(t, gitignore, "codeship.aes")
	dockerignorebits, err := ioutil.ReadFile(".dockerignore")
	check(err)
	dockerignore := string(dockerignorebits)
	assert.Contains(t, dockerignore, "codeship.env")
	assert.Contains(t, dockerignore, "codeship.aes")

	//clean up files
	err = os.Remove(".gitignore")
	check(err)
	err = os.Remove(".dockerignore")
	check(err)
}

func assertArtifact(t *testing.T, artifacts []*BuildArtifact, filePath string) *BuildArtifact {
	var result *BuildArtifact
	for _, artifact := range artifacts {
		if artifact.FilePath == filePath {
			result = artifact
			break
		}
	}
	assert.NotNil(t, result, "expecting "+filePath)
	return result
}
