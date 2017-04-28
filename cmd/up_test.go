package cmd

import (
	"log"
	"testing"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/stretchr/testify/assert"
)

func TestUpEnvVarEqualsSign(t *testing.T) {

	yaml := `
version: "2"
services:
  container:
    image: helloworld
    environment:
      CHAR_EQUAL: foo=bar	
`
	bytes := [][]byte{[]byte(yaml)}

	//use libcompose to parse compose yml
	dockerComposeProject, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeBytes: bytes,
		},
	}, nil)

	if err != nil {
		log.Fatal(err)
	}

	proj, _ := dockerComposeProject.GetServiceConfig("container")

	assert.Equal(t, "foo=bar", proj.Environment.ToMap()["CHAR_EQUAL"])
}
