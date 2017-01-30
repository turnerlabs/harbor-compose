package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/spf13/cobra"
)

// catalogCmd represents the up command
var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Add containers in docker-compose to Harbor Catalog",
	Run:   catalog,
}

func init() {
	RootCmd.AddCommand(catalogCmd)
}

func catalog(cmd *cobra.Command, args []string) {

	// read the harbor compose file
	harborCompose := DeserializeHarborCompose(HarborComposeFile)

	// read the docker compose file
	dockerCompose := DeserializeDockerCompose(DockerComposeFile)

	// use libcompose to parse compose yml file as well (since it supports the full spec)
	dockerComposeProject, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeFiles: []string{DockerComposeFile},
		},
	}, nil)

	if err != nil {
		log.Fatal(err)
	}

	//iterate shipments
	for shipmentName, shipment := range harborCompose.Shipments {
		fmt.Printf("Starting %v ...\n", shipmentName)

		if Verbose {
			log.Printf("cataloging shipment: %v/%v", shipmentName, shipment.Env)
		}

		// catalog the contianers in the shipment
		CatalogContainers(dockerCompose, dockerComposeProject)

		fmt.Println("done")

	} //shipments
}

// CatalogContainers will take a dockerCompose object and catalog all contianers it gets
func CatalogContainers(dockerCompose DockerCompose, dockerComposeProject project.APIProject) {

	if Verbose {
		log.Println("cataloging containers for shipment")
	}

	// loop over containers in docker-compose file
	for containerName, container := range dockerCompose.Services {

		if Verbose {
			log.Printf("processing container: %v", containerName)
		}

		if container.Image == "" {
			log.Fatalln("'image' is required in docker compose file")
		}

		//parse image:tag and map to name/version
		parsedImage := strings.Split(container.Image, ":")

		newContainer := CatalogitContainer{
			Name:    containerName,
			Image:   container.Image,
			Version: parsedImage[1],
		}

		// send POST to catalogit
		// if post failes and says image already exists, do not exit 1
		//trigger shipment
		message, err := Catalogit(newContainer)

		if err != nil {
			log.Fatal(err)
		}

		// cataloged successfully
		if Verbose {
			fmt.Println(message)
		}
	}

	fmt.Println("Containers were successfully cataloged.")
}
