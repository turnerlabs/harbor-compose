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

	yaml "gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

//tests generating a docker-compose.yml from an existing shipment
func TestTransformShipmentToDockerCompose(t *testing.T) {

	//define a ShipmentEnvironment
	shipmentJSON := `
{
  "enableMonitoring": true,
  "name": "dev",
  "parentShipment": {
    "name": "mss-poc-app",
    "group": "mss",
    "envVars": [
      {
        "type": "basic",
        "value": "adds",
        "name": "CUSTOMER"
      },
      {
        "type": "basic",
        "value": "mss-poc-app",
        "name": "PRODUCT"
      },
      {
        "type": "basic",
        "value": "mss-poc-app",
        "name": "PROJECT"
      },
      {
        "type": "basic",
        "value": "mss",
        "name": "PROPERTY"
      }
    ]
  },
  "envVars": [
    {
      "type": "basic",
      "value": "bar",
      "name": "FOO"
    }
  ],
  "providers": [
    {
      "replicas": 2,
      "barge": "corp-sandbox",
      "name": "ec2",
      "envVars": []
    }
  ],
  "containers": [
    {
      "image": "${image}",
      "name": "${service}",
      "envVars": [
        {
          "type": "basic",
          "value": "${healthCheck}",
          "name": "HEALTHCHECK"
        }
      ],
      "ports": [
        {
          "protocol": "http",
          "healthcheck": "${healthCheck}",
          "external": true,
          "primary": true,
          "public_vip": false,
          "enable_proxy_protocol": false,
          "ssl_arn": "",
          "ssl_management_type": "iam",
          "healthcheck_timeout": 1,
          "public_port": ${publicPort},
          "value": ${containerPort},
          "name": "PORT"
        }
      ]
    }
  ]
}	
`

	//update json with test values
	service := "mss-poc-app"
	image := "quay.io/turner/mss-poc-app:1.0.0"
	publicPort := "80"
	containerPort := "3000"
	healthCheck := "/hc"
	foo := "bar"

	shipmentJSON = strings.Replace(shipmentJSON, "${service}", service, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${image}", image, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${publicPort}", publicPort, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${containerPort}", containerPort, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${healthCheck}", healthCheck, 1)

	//deserialize json
	var shipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipment)
	if err != nil {
		log.Fatal(err)
	}

	//test
	dockerCompose := transformShipmentToDockerCompose(&shipment)

	//debug
	data, _ := yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//assertions
	assert.Equal(t, 1, len(dockerCompose.Services))
	composeService := dockerCompose.Services[service]
	assert.NotNil(t, composeService)
	assert.Equal(t, image, composeService.Image)
	assert.Equal(t, fmt.Sprintf("%v:%v", publicPort, containerPort), composeService.Ports[0])
	assert.Equal(t, containerPort, composeService.Environment["PORT"])
	assert.Equal(t, healthCheck, composeService.Environment["HEALTHCHECK"])
	assert.Equal(t, foo, composeService.Environment["FOO"])
}

//tests generating a docker-compose.yml from an existing shipment with multiple containers
func TestTransformShipmentToDockerComposeMultiContainer(t *testing.T) {

	//define a ShipmentEnvironment
	shipmentJSON := `
{
  "enableMonitoring": true,
  "name": "dev",
  "parentShipment": {
    "name": "mss-poc-app",
    "group": "mss",
    "envVars": [
      {
        "type": "basic",
        "value": "adds",
        "name": "CUSTOMER"
      },
      {
        "type": "basic",
        "value": "mss-poc-app",
        "name": "PRODUCT"
      },
      {
        "type": "basic",
        "value": "mss-poc-app",
        "name": "PROJECT"
      },
      {
        "type": "basic",
        "value": "mss",
        "name": "PROPERTY"
      }
    ]
  },
  "envVars": [
    {
      "type": "basic",
      "value": "bar",
      "name": "FOO"
    }
  ],
  "providers": [
    {
      "replicas": 2,
      "barge": "corp-sandbox",
      "name": "ec2",
      "envVars": []
    }
  ],
  "containers": [
    {
      "image": "${image}",
      "name": "${service}",
      "envVars": [
        {
          "type": "basic",
          "value": "${healthCheck}",
          "name": "HEALTHCHECK"
        }
      ],
      "ports": [
        {
          "protocol": "http",
          "healthcheck": "${healthCheck}",
          "external": true,
          "primary": true,
          "public_vip": false,
          "enable_proxy_protocol": false,
          "ssl_arn": "",
          "ssl_management_type": "iam",
          "public_port": ${publicPort},
          "value": ${containerPort},
          "name": "PORT"
        }
      ]
    },
    {
      "image": "${image}",
      "name": "${service2}",
      "envVars": [
        {
          "type": "basic",
          "value": "${healthCheck}",
          "name": "HEALTHCHECK"
        }
      ],
      "ports": [
        {
          "protocol": "http",
          "healthcheck": "${healthCheck}",
          "external": true,
          "primary": true,
          "public_vip": false,
          "enable_proxy_protocol": false,
          "ssl_arn": "",
          "ssl_management_type": "iam",
          "public_port": ${publicPort},
          "value": ${containerPort},
          "name": "PORT"
        }
      ]
    }    
  ]
}	
`
	//update json with test values
	service := "mss-poc-app"
	service2 := "mss-poc-app2"
	image := "quay.io/turner/mss-poc-app:1.0.0"
	publicPort := "80"
	containerPort := "3000"
	healthCheck := "/hc"
	foo := "bar"

	shipmentJSON = strings.Replace(shipmentJSON, "${service}", service, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${service2}", service2, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${image}", image, 2)
	shipmentJSON = strings.Replace(shipmentJSON, "${publicPort}", publicPort, 2)
	shipmentJSON = strings.Replace(shipmentJSON, "${containerPort}", containerPort, 2)
	shipmentJSON = strings.Replace(shipmentJSON, "${healthCheck}", healthCheck, 4)

	//deserialize json
	var shipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipment)
	if err != nil {
		log.Fatal(err)
	}

	//test
	dockerCompose := transformShipmentToDockerCompose(&shipment)

	//debug
	data, _ := yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//assertions
	assert.Equal(t, 2, len(dockerCompose.Services))

	composeService := dockerCompose.Services[service]
	assert.NotNil(t, composeService)
	assert.Equal(t, image, composeService.Image)
	assert.Equal(t, fmt.Sprintf("%v:%v", publicPort, containerPort), composeService.Ports[0])
	assert.Equal(t, containerPort, composeService.Environment["PORT"])
	assert.Equal(t, healthCheck, composeService.Environment["HEALTHCHECK"])
	assert.Equal(t, foo, composeService.Environment["FOO"])

	composeService2 := dockerCompose.Services[service2]
	assert.NotNil(t, composeService2)
	assert.Equal(t, image, composeService2.Image)
	assert.Equal(t, fmt.Sprintf("%v:%v", publicPort, containerPort), composeService2.Ports[0])
	assert.Equal(t, containerPort, composeService2.Environment["PORT"])
	assert.Equal(t, healthCheck, composeService2.Environment["HEALTHCHECK"])
	assert.Equal(t, foo, composeService2.Environment["FOO"])
}

//tests generating a harbor-compose.yml from an existing shipment
func TestTransformShipmentToHarborCompose(t *testing.T) {
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

	//assertions
	assert.Equal(t, 1, len(harborCompose.Shipments))
	composeShipment := harborCompose.Shipments[name]
	assert.NotNil(t, composeShipment)
	assert.Equal(t, group, composeShipment.Group)
	assert.Equal(t, barge, composeShipment.Barge)
	assert.Equal(t, env, composeShipment.Env)
	assert.Equal(t, replicas, composeShipment.Replicas)
	assert.Equal(t, project, composeShipment.Project)
	assert.Equal(t, property, composeShipment.Property)
	assert.Equal(t, product, composeShipment.Product)
	assert.Equal(t, 1, len(composeShipment.Containers))

	//IgnoreImageVersion should default to false
	assert.Equal(t, false, composeShipment.IgnoreImageVersion)

	//both container-level and env-level shipit envvars should get added to docker-compose and not harbor-compose
	assert.Equal(t, containerLevel, dockerCompose.Services["web"].Environment[containerLevel])
	assert.Equal(t, envLevel, dockerCompose.Services["web"].Environment[envLevel])
	assert.NotEqual(t, envLevel, composeShipment.Environment[envLevel])
}

func getSampleShipmentJSON() string {
	return `
{
  "name": "${env}",
  "parentShipment": {
    "name": "${name}",
    "group": "${group}",
    "envVars": [
      {
        "type": "basic",
        "value": "customer",
        "name": "CUSTOMER"
      },
      {
        "type": "basic",
        "value": "${product}",
        "name": "PRODUCT"
      },
      {
        "type": "basic",
        "value": "${project}",
        "name": "PROJECT"
      },
      {
        "type": "basic",
        "value": "${property}",
        "name": "PROPERTY"
      }
    ]
  },
  "envVars": [
    {
      "type": "basic",
      "value": "${envLevel}",
      "name": "ENV_LEVEL"
    }
  ],
  "providers": [
    {
      "replicas": ${replicas},
      "barge": "${barge}",
      "name": "ec2",
      "envVars": []
    }
  ],
  "containers": [
    {
      "image": "quay.io/turner/web:1.0",
      "name": "${container}",
      "envVars": [
        {
          "type": "basic",
          "value": "/hc",
          "name": "HEALTHCHECK"
        },
        {
          "type": "basic",
          "value": "${containerLevel}",
          "name": "CONTAINER_LEVEL"
        }        
      ],
      "ports": [
        {
          "protocol": "http",
          "healthcheck": "/hc",
          "external": true,
          "primary": true,
          "public_vip": false,
          "enable_proxy_protocol": false,
          "ssl_arn": "",
          "ssl_management_type": "iam",
          "healthcheck_timeout": 1,
          "public_port": 80,
          "value": 5000,
          "name": "PORT"
        }
      ]
    }
  ]
}	
`
}

//tests the local build provider
func TestTransformShipmentToDockerComposeBuildProviderLocal(t *testing.T) {
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

	//load local build provider
	provider, err := getBuildProvider("local")
	if err != nil {
		t.Fail()
	}

	//run the build provider
	_, err = provider.ProvideArtifacts(&dockerCompose, &harborCompose, "token")
	if err != nil {
		t.Fail()
	}

	//debug
	data, _ = yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//docker compose configuration should have the build directive
	assert.NotEmpty(t, svc.Build)
}

//tests generate --build-provider circleciv1
func TestTransformShipmentToDockerComposeBuildProviderCircleCI(t *testing.T) {
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
	provider, err := getBuildProvider("circleciv1")
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

	//the provider should output a circle.yml
	assert.NotNil(t, artifacts)
	assert.Equal(t, "circle.yml", artifacts[0].FilePath)
}

//tests generate --build-provider circleciv2
func TestTransformShipmentToDockerComposeBuildProviderCircleCIv2(t *testing.T) {
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

//tests generate --build-provider codeship
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
	check(err)

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
	assertArtifact(t, artifacts, "docker-push.sh")

	//assert that codeship.env is added to .gitignore
	gitignore, err := ioutil.ReadFile(".gitignore")
	check(err)
	assert.Contains(t, string(gitignore), "codeship.env")
	err = os.Remove(".gitignore")
	check(err)
}

func assertArtifact(t *testing.T, artifacts []*BuildArtifact, filePath string) {
	found := false
	for _, artifact := range artifacts {
		if artifact.FilePath == filePath {
			found = true
			break
		}
	}
	assert.True(t, found, "expecting "+filePath)
}
