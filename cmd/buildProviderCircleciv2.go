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
  var imageRepo, imageVersion string
	for _, svc := range dockerCompose.Services {

		//set build configuration to string containing a path to the build context
		svc.Build = "."
		if platform == "ecsfargate" {
			svc.Build = "../"
    }
    
    //split out parts of docker image for use later
    imageRepo, imageVersion, _ = parseDockerImage(svc.Image)

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

    //output a config.env
		if harborCompose != nil {
			for name, shipment := range harborCompose.Shipments {
				artifacts = append(artifacts, createArtifact(".circleci/config.env", getConfigEnv(name, shipment.Env, imageRepo, imageVersion)))
				break
			}
		}
	}
	fmt.Println()

	return artifacts, nil
}

//parse a docker image into it's constituent parts: ${repo}:${version}-${prerelease}
func parseDockerImage(image string) (string, string, string) {
  repoVersion := strings.Split(image, ":")
  version := "latest"
  pre := ""
  if len(repoVersion) > 1 {
    versionPrerelease := strings.Split(repoVersion[1], "-")
    version = versionPrerelease[0]
    if version == "" {
      version = "latest"
    }
    if len(versionPrerelease) > 0 {
      pre = strings.Join(versionPrerelease[1:], "-")
    }
  }
  return repoVersion[0], version, pre
}

func getConfigEnv(app string, env string, repo string, version string) string {
  service := fmt.Sprintf("%s-%s", app, env)
  result := fmt.Sprintf(`
export FARGATE_CLUSTER=%s
export FARGATE_SERVICE=%s
export REPO=%s
export VERSION=%s
 `, service, service, repo, version)

  return result
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
  template := `
version: 2
jobs:
  build:
    docker:
      - image: quay.io/turner/fargate-cicd
    environment:
      VAR: .circleci/config.env
    steps:
      - checkout
      - setup_remote_docker:
          version: 17.11.0-ce          
      - run:
          name: Set docker image
          command: |
            source ${VAR}
            # for node.js apps you can use version from package.json
            # VERSION=$(jq -r .version < package.json)
            BUILD=${CIRCLE_BUILD_NUM}
            if [ "${CIRCLE_BRANCH}" != "master" ]; then
              BUILD=${CIRCLE_BRANCH}.${CIRCLE_BUILD_NUM}
            fi
            echo "export IMAGE=${REPO}:${VERSION}-${BUILD}" >> ${VAR}
            cat ${VAR}
      - run:        
          name: Login to registry
          command: eval $(aws ecr get-login --no-include-email)
      - run:
          name: Build app image
          command: . ${VAR}; docker build -t ${IMAGE} .
      - run:
          name: Push app image to registry
          command: . ${VAR}; docker push ${IMAGE}
      - run:
          name: Deploy
          command: . ${VAR}; fargate service deploy -i ${IMAGE}`

	return template
}
