package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

// catalogCmd represents the up command
var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Add containers defined in a docker compose file to the harbor catalog",
	Run:   catalog,
}

func init() {
	RootCmd.AddCommand(catalogCmd)
}

// catalog is called by CLI catalog function. Reads from
// dockerCompose and harborCompose and loops over containers to
// catalog images
func catalog(cmd *cobra.Command, args []string) {

	// read the harbor compose file
	harborCompose := DeserializeHarborCompose(HarborComposeFile)

	// read the docker compose file
	dockerCompose := DeserializeDockerCompose(DockerComposeFile)

	//iterate shipments
	for shipmentName, shipment := range harborCompose.Shipments {
		fmt.Printf("Cataloging images for Shipment: %v ...\n", shipmentName)

		if Verbose {
			log.Printf("cataloging shipment: %v/%v", shipmentName, shipment.Env)
		}

		// loop over containers in docker-compose file
		for containerIndex, containerName := range shipment.Containers {

			if Verbose {
				log.Printf("processing container: %v %v", containerName, containerIndex)
			}

			//lookup the container in the list of services in the docker-compose file
			container, isset := dockerCompose.Services[containerName]

			if isset == true {
				// catalog the containers in the shipment
				CatalogContainer(containerName, container.Image)
			}
		}

		fmt.Println("done")

	} //shipments
}

// CatalogContainer will create a container object from name and image params
// Will send POST to catalogit api
func CatalogContainer(name string, image string) {

	if Verbose {
		log.Printf("cataloging container %v", name)
	}

	//parse image:tag and map to name/version
	parsedImage := strings.Split(image, ":")

	newContainer := CatalogitContainer{
		Name:    name,
		Image:   image,
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
