package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:    "restart",
	Short:  "Restart your application",
	Long:   `Adds a dummy variable to the shipment and triggers it.`,
	Run:    restart,
	PreRun: preRunHook,
}

func init() {
	RootCmd.AddCommand(restartCmd)
}

// restart your shipment
func restart(cmd *cobra.Command, args []string) {

	//make sure user is authenticated
	username, token, err := Login()
	check(err)

	//read the harbor compose file
	harborCompose := DeserializeHarborCompose(HarborComposeFile)

	//iterate shipments
	for shipmentName, shipment := range harborCompose.Shipments {
		fmt.Printf("Restarting %v %v ...\n", shipmentName, shipment.Env)

		t := time.Now()
		time := username + "_" + t.Format("20060102150405")

		envVar := EnvVarPayload{
			Name:  envVarNameRestart,
			Value: time,
			Type:  "basic",
		}

		//update env var
		SaveEnvVar(username, token, shipmentName, shipment, envVar, shipment.Containers[0])

		//trigger
		Trigger(shipmentName, shipment.Env)

		fmt.Println("done")

	} //shipments
}
