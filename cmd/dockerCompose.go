package cmd

import (
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

	return &dockerCompose
}
