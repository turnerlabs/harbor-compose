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

// deployCmd represents the up command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Trigger an image deployment for all shipments and containers defined in compose files",
	Long: `Trigger an image deployment for all shipments and containers defined in compose files.
	
Note that the deploy command is a subset of the up command without updates for environment variables, replicas, barge info, etc.

Also note that a shipment build token is required to be specified as an environment variable using the specific naming convention below.  Shipment build tokens are generated at the environment level so you can use any environment you wish.

Example (shipment = mss-app-web):

MSS_APP_WEB_DEV_TOKEN=xyz harbor-compose deploy
`,
	Run: deploy,
}

var environmentOverride string

func init() {
	deployCmd.PersistentFlags().StringVarP(&environmentOverride, "env", "e", "", "override the shipment environment specified in the harbor compose file.")
	RootCmd.AddCommand(deployCmd)
}

//deploy iterates shipments and containers and uses the Customs API to trigger deployments.
func deploy(cmd *cobra.Command, args []string) {

	//read the harbor compose file
	harborCompose := DeserializeHarborCompose(HarborComposeFile)

	//use libcompose to parse yml file
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
		fmt.Printf("deploying images for shipment: %v %v ...\n", shipmentName, shipment.Env)

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

			//allow --env flag to override environment specified in compose file
			shipmentEnv := shipment.Env
			if len(environmentOverride) > 0 {
				shipmentEnv = environmentOverride
			}

			//get for envvar for this shipment/environment
			buildTokenEnvVar := getBuildTokenEnvVar(shipmentName, shipmentEnv)

			Deploy(shipmentName, shipmentEnv, buildTokenEnvVar, deployRequest, "ec2")
		}

		fmt.Println("done")

	} //shipments
}
