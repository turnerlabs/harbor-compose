package cmd

import (
	"fmt"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
)

//unmarshal docker compose yaml string into a compose APIProject
func unmarshalDockerCompose(yamlString string) project.APIProject {
	if Verbose {
		log.Printf("unmarshalDockerCompose - %v", yamlString)
	}

	yamlBits := []byte(yamlString)

	//use libcompose to parse compose yml
	bytes := [][]byte{yamlBits}
	dockerCompose, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeBytes: bytes,
			ProjectName:  "required",
		},
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	return dockerCompose
}

//unmarshal harbor compose yaml
func unmarshalHarborCompose(yamlString string) HarborCompose {
	yamlBits := []byte(yamlString)
	var harborCompose HarborCompose
	err := yaml.Unmarshal(yamlBits, &harborCompose)
	if err != nil {
		log.Fatalf("harbor compose error: %v", err)
	}
	return harborCompose
}

//unmarshals both docker compose and harbor compose yaml strings
func unmarshalCompose(dockerComposeYaml string, harborComposeYaml string) (project.APIProject, HarborCompose) {
	dc := unmarshalDockerCompose(dockerComposeYaml)
	hc := unmarshalHarborCompose(harborComposeYaml)
	return dc, hc
}

//unmarshals both docker compose and harbor compose yaml files
func unmarshalComposeFiles(dockerComposeFile string, harborComposeFile string) (project.APIProject, HarborCompose) {
	dc := DeserializeDockerCompose(dockerComposeFile)
	hc := DeserializeHarborCompose(harborComposeFile)
	return dc, hc
}

func getDockerComposeService(dockerCompose project.APIProject, container string) *config.ServiceConfig {
	serviceConfig, success := dockerCompose.GetServiceConfig(container)
	if !success {
		fmt.Printf("ERROR: Container: %v defined in %v cannot be found in %v\n", container, HarborComposeFile, DockerComposeFile)
		os.Exit(-1)
	}
	return serviceConfig
}
