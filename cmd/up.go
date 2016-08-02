package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start your application",
	Long:  ``,
	Run:   up,
}

func init() {
	RootCmd.AddCommand(upCmd)
}

func up(cmd *cobra.Command, args []string) {

	//read the harbor compose file
	harborCompose := DeserializeHarborCompose(HarborComposeFile)

	//read the docker compose file
	dockerCompose := DeserializeDockerCompose(DockerComposeFile)

	_, token := Login()

	//iterate shipments
	for shipmentName, shipment := range harborCompose.Shipments {
		fmt.Printf("Starting %v ...\n", shipmentName)

		if Verbose {
			log.Printf("processing shipment: %v/%v", shipmentName, shipment.Env)
		}

		//iterate defined containers and apply container level updates
		for _, container := range shipment.Containers {
			if Verbose {
				log.Printf("processing container: %v", container)
			}

			//lookup the container in the list of services in the docker-compose file
			dockerService := dockerCompose.Services[container]

			//update the shipment/container with the new image
			UpdateContainerImage(token, shipmentName, shipment, container, dockerService)

			//update container-level envvars
			for evName, evValue := range dockerService.Environment {

				if Verbose {
					log.Println("processing " + evName)
				}

				//create an envvar object
				envVarPayload := EnvVarPayload{
					Name:  evName,
					Value: evValue,
					Type:  "basic",
				}

				//save the envvar
				SaveEnvVar(token, shipmentName, shipment, envVarPayload, container)

			} //envvars
		}

		//convert the specified barge into an env var
		if len(shipment.Barge) > 0 {

			//initialize the environment map if it doesn't exist
			if shipment.Environment == nil {
				shipment.Environment = make(map[string]string)
			}

			//set the BARGE env var
			shipment.Environment["BARGE"] = shipment.Barge
		}

		//update shipment/environment-level envvars
		for evName, evValue := range shipment.Environment {

			if Verbose {
				log.Println("processing " + evName)
			}

			//create an envvar object
			envVarPayload := EnvVarPayload{
				Name:  evName,
				Value: evValue,
				Type:  "basic",
			}

			//save the envvar
			SaveEnvVar(token, shipmentName, shipment, envVarPayload, "")

		} //envvars

		//update shipment level configuration
		UpdateShipment(shipmentName, shipment, token)

		//trigger shipment
		result := Trigger(shipmentName, shipment.Env)

		if len(result.Messages) > 0 {
			fmt.Println(result.Messages[0])
		}

		fmt.Println("done")

	} //shipments
}
