package cmd

import (
	"errors"
	"strings"
)

//BuildArtifact represents a build artifact
type BuildArtifact struct {
	FilePath     string
	FileContents string
}

//BuildProvider represents a build provider
type BuildProvider interface {

	//build providers can manipulate the docker compose configuration and output build artifacts
	ProvideArtifacts(dockerCompose *DockerCompose, harborCompose *HarborCompose, token string) ([]*BuildArtifact, error)
}

//return a build provider based on its name
func getBuildProvider(provider string) (BuildProvider, error) {

	if strings.ToLower(provider) == "local" {
		return LocalBuild{}, nil
	}

	if strings.ToLower(provider) == "circleciv1" {
		return CircleCIv1{}, nil
	}

	if strings.ToLower(provider) == "circleciv2" {
		return CircleCIv2{}, nil
	}

	if strings.ToLower(provider) == "codeship" {
		return Codeship{}, nil
	}

	return nil, errors.New("no build provider found for: " + provider)
}
