package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/spf13/cobra"
)

// deployCmd represents the up command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Trigger an image deployment for all shipments and containers defined in compose files",
	Long:  "Note that the deploy command is a subset of the up command without updates for environment variables, replicas, barge info, etc.",
	Run:   deploy,
}

var shipmentBuildToken string

func init() {
	deployCmd.PersistentFlags().StringVarP(&shipmentBuildToken, "token", "t", "", "the shipment build token to use")
	RootCmd.AddCommand(deployCmd)
}

//deploy iterates shipments and containers and uses the Customs API to trigger deployments.
func deploy(cmd *cobra.Command, args []string) {

	//obtain build token
	buildTokenEnvVar := os.Getenv("BUILD_TOKEN")

	//cli flag overrides env var
	if len(shipmentBuildToken) == 0 {
		shipmentBuildToken = buildTokenEnvVar
	}

	//validate build token
	if len(shipmentBuildToken) == 0 {
		log.Fatal("A shipment build token is required. Please specify an environment variable named, \"BUILD_TOKEN\" or the --token flag.")
	}

	//read the harbor compose file
	harborCompose := DeserializeHarborCompose(HarborComposeFile)

	//use libcompose to parse compose yml file
	dockerComposeProject, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeFiles: []string{DockerComposeFile},
		},
	}, nil)

	if err != nil {
		log.Fatal("error parsing compose file" + err.Error())
	}

	//validate the compose file
	_, err = dockerComposeProject.Config()
	if err != nil {
		log.Fatal("error parsing compose file" + err.Error())
	}

	//iterate shipments
	for shipmentName, shipment := range harborCompose.Shipments {
		fmt.Printf("deploying images for Shipment: %v %v ...\n", shipmentName, shipment.Env)

		// loop over containers in docker-compose file
		for _, containerName := range shipment.Containers {

			//lookup the container in the list of services in the docker-compose file
			serviceConfig, found := dockerComposeProject.GetServiceConfig(containerName)
			if !found {
				log.Fatal("could not find service in docker compose file")
			}

			//parse image:tag and map to name/version
			parsedImage := strings.Split(serviceConfig.Image, ":")
			tag := parsedImage[1]

			//lookup container image in the catalog and catalog if missing
			catalog := !IsContainerVersionCataloged(containerName, tag)

			//now deploy container
			if Verbose {
				log.Printf("deploying container: %v\n", containerName)
			}

			deployRequest := DeployRequest{
				Name:    containerName,
				Image:   serviceConfig.Image,
				Version: tag,
				Catalog: catalog,
			}

			Deploy(shipmentName, shipment.Env, shipmentBuildToken, deployRequest)

		}

		fmt.Println("done")

	} //shipments
}
