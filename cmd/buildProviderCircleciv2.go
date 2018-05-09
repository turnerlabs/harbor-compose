package cmd

import (
	"fmt"
	"strings"
)

//CircleCIv2 represents a Circle CI build provider (v2)
type CircleCIv2 struct{}

//ProvideArtifacts -
func (provider CircleCIv2) ProvideArtifacts(dockerCompose *DockerCompose, harborCompose *HarborCompose, token string, platform string) ([]*BuildArtifact, error) {

	if !(platform == "harbor" || platform == "ecsfargate") {
		return nil, fmt.Errorf("build provider doesn't support platform %s", platform)
	}

	//iterate containers
	for _, svc := range dockerCompose.Services {

		//set build configuration to string containing a path to the build context
		svc.Build = "."
		if platform == "ecsfargate" {
			svc.Build = "../"
		}

		//add the circle ci build number to the image tag
		svc.Image += "-${CIRCLE_BUILD_NUM}"

		//remove environment variables and ports since they're not needed for ci/cd
		svc.Environment = nil
		svc.EnvFile = nil
		svc.Ports = nil
	}

	//output circle.yml
	artifacts := []*BuildArtifact{}
	artifacts = append(artifacts, createArtifact(".circleci/config.yml", getCircleCIv2YAML(platform)))

	fmt.Println()
	if platform == "harbor" {

		fmt.Println("Be sure to supply the following environment variables in your Circle CI build:\nDOCKER_USER (quay.io registry user, e.g., turner+my_team)\nDOCKER_PASS (quay.io registry password)")
		if harborCompose != nil {
			for name, shipment := range harborCompose.Shipments {
				fmt.Print(getBuildTokenName(name, shipment.Env))
				fmt.Print("=")
				fmt.Println(token)
			}
		}
	} else if platform == "ecsfargate" {

		fmt.Println(`Be sure to supply the following environment variables in your Circle CI build:
AWS_ACCESS_KEY_ID (terraform output cicd_keys)
AWS_SECRET_ACCESS_KEY (terraform output cicd_keys)
AWS_DEFAULT_REGION=us-east-1
`)

		//output a fargate.yml (for fargate cli)
		if harborCompose != nil {
			for name, shipment := range harborCompose.Shipments {
				artifacts = append(artifacts, createArtifact(".circleci/fargate.yml", getFargateYaml(name, shipment.Env)))
				break
			}
		}

		//output a docker-compose.yml (referenced by circle yaml)
		dockerComposeYaml := marshalDockerCompose(*dockerCompose)
		artifacts = append(artifacts, createArtifact(".circleci/docker-compose.yml", string(dockerComposeYaml)))
	}
	fmt.Println()

	return artifacts, nil
}

func getCircleCIv2YAML(platform string) string {
	if platform == "harbor" {
		return getCircleCIv2YAMLForHarbor()
	} else if platform == "ecsfargate" {
		return getCircleCIv2YAMLForEcsFargate()
	}
	return ""
}

func getCircleCIv2YAMLForHarbor() string {
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

func getCircleCIv2YAMLForEcsFargate() string {
	template := `version: 2
  jobs:
    build:
      docker:
        - image: quay.io/turner/fargate-cicd
      steps:
        - checkout
        - setup_remote_docker:
            version: 17.11.0-ce
        - run:
            name: Build app image
            command: docker-compose build
            working_directory: ~/project/.circleci
        - run:        
            name: Login to registry
            command: eval $(aws ecr get-login --no-include-email)
        - run:
            name: Push app image to registry
            command: docker-compose push
            working_directory: ~/project/.circleci
        - run:
            name: Deploy develop branch to fargate
            command: |
              if [ "${CIRCLE_BRANCH}" == "develop" ]; then 
                fargate service deploy -f docker-compose.yml
              fi
            working_directory: ~/project/.circleci`

	return template
}
