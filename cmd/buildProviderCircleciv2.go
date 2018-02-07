package cmd

import (
	"fmt"
	"strings"
)

//CircleCIv2 represents a Circle CI build provider (v2)
type CircleCIv2 struct{}

//ProvideArtifacts -
func (provider CircleCIv2) ProvideArtifacts(dockerCompose *DockerCompose, harborCompose *HarborCompose, token string) ([]*BuildArtifact, error) {

	//iterate containers
	for _, svc := range dockerCompose.Services {

		//set build configuration to string containing a path to the build context
		svc.Build = "."

		//add the circle ci build number to the image tag
		svc.Image += "-${CIRCLE_BUILD_NUM}"

		//remove environment variables and ports since they're not needed for ci/cd
		svc.Environment = nil
		svc.Ports = nil
	}

	//output circle.yml
	artifacts := []*BuildArtifact{}
	artifacts = append(artifacts, createArtifact(".circleci/config.yml", getCircleCIv2YAML()))

	fmt.Println()
	fmt.Println("Be sure to supply the following environment variables in your Circle CI build:\nDOCKER_USER (quay.io registry user, e.g., turner+my_team)\nDOCKER_PASS (quay.io registry password)")
	if harborCompose != nil {
		for name, shipment := range harborCompose.Shipments {
			fmt.Print(getBuildTokenName(name, shipment.Env))
			fmt.Print("=")
			fmt.Println(token)
		}
	}
	fmt.Println()

	return artifacts, nil
}

func getCircleCIv2YAML() string {
	template := `
version: 2
jobs:
  build:
    docker:
      - image: quay.io/turner/harbor-cicd-image:${HC_VERSION}
    working_directory: ~/app
    steps:
      - checkout
      - setup_remote_docker:
          version: 17.06.0-ce
      - run:
          name: Build app image
          command: docker-compose build
      - run:        
          name: Login to registry
          command: docker login -u="${DOCKER_USER}" -p="${DOCKER_PASS}" quay.io
      - run:
          name: Push app image to registry
          command: docker-compose push
      - run:
          name: Catalog in Harbor
          command: harbor-compose catalog
      - run:
          name: Deploy to Harbor
          command: |
            if [ "${CIRCLE_BRANCH}" == "develop" ]; then 
              harbor-compose deploy;
            fi`

	return strings.Replace(template, "${HC_VERSION}", Version, 1)
}
