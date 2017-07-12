package cmd

import (
	"log"

	yaml "gopkg.in/yaml.v2"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
)

//unmarshal docker compose yaml
func unmarshalDockerCompose(yamlString string) (DockerCompose, project.APIProject) {
	if Verbose {
		log.Printf("unmarshalDockerCompose - %v", yamlString)
	}

	yamlBits := []byte(yamlString)

	//parse the docker compose file (only used for writing used by generate)
	var dockerCompose DockerCompose
	err := yaml.Unmarshal(yamlBits, &dockerCompose)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	//use libcompose to parse compose yml
	bytes := [][]byte{yamlBits}
	dockerComposeProject, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeBytes: bytes,
			ProjectName:  "required",
		},
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	return dockerCompose, dockerComposeProject
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

//unmarshals both docker compose and harbor compose yaml
func unmarshalCompose(dockerComposeYaml string, harborComposeYaml string) (project.APIProject, HarborCompose) {
	_, dc := unmarshalDockerCompose(dockerComposeYaml)
	hc := unmarshalHarborCompose(harborComposeYaml)
	return dc, hc
}
