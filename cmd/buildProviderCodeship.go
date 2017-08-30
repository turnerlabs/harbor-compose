package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

//Codeship represents a Codeship build provider
type Codeship struct{}

//ProvideArtifacts -
func (provider Codeship) ProvideArtifacts(dockerCompose *DockerCompose, harborCompose *HarborCompose, token string) ([]*BuildArtifact, error) {

	//iterate containers
	for _, svc := range dockerCompose.Services {

		//set build configuration to string containing a path to the build context
		svc.Build = "."

		//add the build number to the image tag
		svc.Image += "-${CI_BUILD_ID}"

		//remove environment variables and ports since they're not needed for ci/cd
		svc.Environment = nil
		svc.Ports = nil
	}

	//add artificats
	artifacts := []*BuildArtifact{}
	artifacts = append(artifacts, createArtifact("codeship-services.yml", getCodeshipServices()))
	artifacts = append(artifacts, createArtifact("codeship-steps.yml", getCodeshipSteps()))
	artifacts = append(artifacts, createArtifact("codeship.env", getCodeshipEnv(harborCompose, token)))
	artifacts = append(artifacts, createArtifact("codeship.aes", ""))

	//add an executable script
	artifacts = append(artifacts, &BuildArtifact{
		FilePath:     "docker-push.sh",
		FileContents: getDockerPush(),
		FileMode:     0777,
	})

	//look for .gitignore
	if _, err := os.Stat(".gitignore"); err == nil {
		//update .gitignore
		file, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_WRONLY, 0600)
		check(err)
		defer file.Close()
		_, err = file.WriteString("\ncodeship.env")
		check(err)
		_, err = file.WriteString("\ncodeship.aes")
		check(err)

	} else {
		//doesn't exist, create it
		err := ioutil.WriteFile(".gitignore", []byte("codeship.env\ncodeship.aes"), 0644)
		check(err)
	}

	fmt.Println()
	fmt.Println(`Now you just need to:

- add your quay.io registry credentials to codeship.env
- download your AES key from your codeship project and put it in codeship.aes
- encrypt your codeship.env by running 'jet encrypt codeship.env codeship.env.encrypted'
- check in codeship.env.encrypted but don't check in codeship.env`)
	fmt.Println()

	return artifacts, nil
}

func getDockerPush() string {
	template := `#!/bin/bash
set -e

# note that this script is only required until codeship
# supports properly evaluated envvars in codeship.steps.yml
docker login -u="${DOCKER_USER}" -p="${DOCKER_PASS}" quay.io  
docker-compose push`

	return strings.Replace(template, "${HC_VERSION}", Version, 1)
}

func getCodeshipServices() string {
	template := `cicd:  
  image: quay.io/turner/harbor-cicd-image:${HC_VERSION}
  encrypted_env_file: codeship.env.encrypted
  add_docker: true
  volumes:
  - ./:/app`

	return strings.Replace(template, "${HC_VERSION}", Version, 1)
}

func getCodeshipSteps() string {
	template := `- service: cicd
  name: build image
  command: docker-compose build

- service: cicd
  name: push image to registry
  command: ./docker-push.sh

- service: cicd
  name: catalog image in harbor
  command: harbor-compose catalog

- service: cicd
  tag: develop
  name: deploy develop branch to harbor
  command: harbor-compose deploy`

	return template
}

func getCodeshipEnv(harborCompose *HarborCompose, token string) string {
	template := `DOCKER_USER=xyz
DOCKER_PASS=xyz
${SHIPMENT_BUILD_TOKEN_NAME}=${SHIPMENT_BUILD_TOKEN_VALUE}
`
	name, shipment := getFirstShipment(harborCompose)
	template = strings.Replace(template, "${SHIPMENT_BUILD_TOKEN_NAME}", getBuildTokenName(name, shipment.Env), 1)
	template = strings.Replace(template, "${SHIPMENT_BUILD_TOKEN_VALUE}", token, 1)

	return template
}

func getFirstShipment(harborCompose *HarborCompose) (string, *ComposeShipment) {
	var shipmentName string
	var shipment ComposeShipment
	for name, s := range harborCompose.Shipments {
		shipmentName = name
		shipment = s
		break
	}
	return shipmentName, &shipment
}
