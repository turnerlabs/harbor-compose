package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// DeserializeDockerCompose deserializes a docker-compose.yml file into an object
func DeserializeDockerCompose(file string) *DockerCompose {

	//read the harbor compose file
	dockerComposeData, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	//parse the harbor compose file
	var dockerCompose DockerCompose
	err = yaml.Unmarshal([]byte(dockerComposeData), &dockerCompose)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if dockerCompose.Version != "2" {
		log.Fatal("only docker-compose format v2 is supported")
	}

	return &dockerCompose
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
