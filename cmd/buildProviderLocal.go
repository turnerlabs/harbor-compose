package cmd

//LocalBuild represents a local build provider
type LocalBuild struct{}

//ProvideArtifacts -
func (provider LocalBuild) ProvideArtifacts(dockerCompose *DockerCompose, harborCompose *HarborCompose, token string, platform string) ([]*BuildArtifact, error) {
	//set build configuration to string containing a path to the build context
	for _, svc := range dockerCompose.Services {
		svc.Build = "."
	}
	return nil, nil
}
