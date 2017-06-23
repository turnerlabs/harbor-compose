package cmd

import (
	"fmt"
	"strings"
)

//CircleCIv1 represents a Circle CI build provider
type CircleCIv1 struct{}

//ProvideArtifacts -
func (provider CircleCIv1) ProvideArtifacts(dockerCompose *DockerCompose, harborCompose *HarborCompose, token string) ([]*BuildArtifact, error) {

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
	artifacts = append(artifacts, &BuildArtifact{
		FilePath:     "circle.yml",
		FileContents: getCircleCIYAML(),
	})

	fmt.Println()
	fmt.Println("Be sure to supply the following environment variables in your Circle CI build:\nDOCKER_USER (registry user)\nDOCKER_PASS (registry password)")
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

func getCircleCIYAML() string {
	template := `
machine:
  pre:
    # install newer docker and docker-compose
    - curl -sSL https://s3.amazonaws.com/circle-downloads/install-circleci-docker.sh | bash -s -- 1.10.0
    - pip install docker-compose==1.11.2

    # install harbor-compose
    - sudo wget -O /usr/local/bin/harbor-compose https://github.com/turnerlabs/harbor-compose/releases/download/${HC_VERSION}/ncd_linux_amd64 && sudo chmod +x /usr/local/bin/harbor-compose
  services:
    - docker

dependencies:
  override:
    # login to quay registry
    - docker login -u="${DOCKER_USER}" -p="${DOCKER_PASS}" -e="." quay.io

compile:
  override:
    - docker-compose build

test:
  override:
    - docker-compose up -d
    - echo "tests run here"
    - docker-compose down

deployment:
  CI:
    branch: master
    commands:
      # push image to registry and catalog in harbor
      - docker-compose push
      - harbor-compose catalog
  CD:
    branch: develop
    commands:
      # push image to registry and deploy to harbor
      - docker-compose push
      - harbor-compose deploy	
	`

	return strings.Replace(template, "${HC_VERSION}", Version, 1)
}
