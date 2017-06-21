package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/docker/libcompose/project"

	"gopkg.in/yaml.v2"
)

// DeserializeDockerCompose deserializes a docker-compose.yml file into an object
func DeserializeDockerCompose(file string) (DockerCompose, project.APIProject) {

	//read the docker compose file from disk
	dockerComposeData, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	//marshal into compose objects
	dockerCompose, dockerComposeProject := unmarshalDockerCompose(string(dockerComposeData))

	if dockerCompose.Version != "2" {
		log.Fatal("only docker-compose format v2 is supported")
	}

	return dockerCompose, dockerComposeProject
}

// SerializeDockerCompose serializes an object to a docker-compose.yml file
func SerializeDockerCompose(dockerCompose DockerCompose, file string) {

	//serialize object to yaml
	data, err := yaml.Marshal(dockerCompose)
	if err != nil {
		log.Fatalf("error marshaling yaml: %v", err)
	}

	if Verbose {
		log.Printf("writing docker-compose file to %v", file)
	}

	//write yaml to docker-compose.yml
	err = ioutil.WriteFile(file, data, 0644)
	if err != nil {
		log.Fatalf("error writing %v: %v", DockerComposeFile, err)
	}

	if Verbose {
		fmt.Println()
		fmt.Printf(string(data))
	}
}
