package cmd

import (
	"fmt"
	"log"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/spf13/cobra"
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart a Shipment",
	Long:  `Adds a dummy variable to the shipment and triggers it.`,
	Run:   restart,
}

func init() {
	RootCmd.AddCommand(restartCmd)
}

// restart your shipment
func restart(cmd *cobra.Command, args []string) {

	//make sure user is authenticated
	username, token, err := Login()
	if err != nil {
		log.Fatal(err)
	}

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
		fmt.Printf("restarting shipment: %v %v ...\n", shipmentName, shipment.Env)
		shipmentEnv := shipment.Env
		Restart(username, token, shipmentName, shipmentEnv)
		fmt.Println("done")

	} //shipments
}
