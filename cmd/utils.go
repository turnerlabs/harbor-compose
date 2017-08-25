package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
)

func check(e error) {
	if e != nil {
		log.Fatal("ERROR: ", e)
	}
}

func getDockerComposeService(dockerCompose project.APIProject, container string) *config.ServiceConfig {
	serviceConfig, success := dockerCompose.GetServiceConfig(container)
	if !success {
		fmt.Printf("ERROR: Container: %v defined in %v cannot be found in %v\n", container, HarborComposeFile, DockerComposeFile)
		os.Exit(-1)
	}
	return serviceConfig
}
