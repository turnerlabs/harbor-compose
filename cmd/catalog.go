package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

// catalogCmd represents the up command
var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Add container images defined in compose files to the Harbor catalog",
	Long: `Add container images defined in compose files to the Harbor catalog.
	
This command is safe to run multiple times.  In other words, the command will not fail if the specified name/version has already been cataloged.

Also note that a shipment build token is required to be specified as an environment variable using the specific naming convention below.  Shipment build tokens are generated at the environment level so you can use any environment you wish.

Example (shipment = mss-app-web):

MSS_APP_WEB_DEV_TOKEN=xyz harbor-compose catalog
`,
	Run:    catalog,
	PreRun: preRunHook,
}

func init() {
	RootCmd.AddCommand(catalogCmd)
}

// catalog is called by CLI catalog function. Reads from
// dockerCompose and harborCompose and loops over containers to
// catalog images
func catalog(cmd *cobra.Command, args []string) {

	//read the compose files
	dockerCompose, harborCompose := unmarshalComposeFiles(DockerComposeFile, HarborComposeFile)

	//validate the compose file
	_, err := dockerCompose.Config()
	if err != nil {
		check(errors.New("error parsing compose file" + err.Error()))
	}

	//iterate shipments
	for shipmentName, shipment := range harborCompose.Shipments {
		fmt.Printf("cataloging images for shipment: %v ...\n", shipmentName)

		// loop over containers in docker-compose file
		for _, containerName := range shipment.Containers {

			//lookup the container in the list of services in the docker-compose file
			serviceConfig, found := dockerCompose.GetServiceConfig(containerName)
			if !found {
				check(errors.New("could not find service in docker compose file"))
			}

			//parse image:tag
			parsedImage := strings.Split(serviceConfig.Image, ":")
			tag := parsedImage[1]

			//lookup container image in the catalog and catalog if missing
			if !IsContainerVersionCataloged(containerName, tag) {

				if Verbose {
					log.Printf("cataloging container: %v\n", containerName)
				}

				catalogContainer := CatalogitContainer{
					Name:    containerName,
					Image:   serviceConfig.Image,
					Version: tag,
				}

				//get for envvar for this shipment/environment
				buildTokenEnvVar := getBuildTokenEnvVar(shipmentName, shipment.Env)

				CatalogCustoms(shipmentName, shipment.Env, buildTokenEnvVar, catalogContainer, "ec2")

			} else if Verbose {
				log.Printf("%s:%s has already been cataloged\n", containerName, tag)
			}
		}

		fmt.Println("done")

	} //shipments
}
