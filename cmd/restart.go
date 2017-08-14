package cmd

import (
	"fmt"
	"log"

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

	//iterate shipments
	for shipmentName, shipment := range harborCompose.Shipments {
		fmt.Printf("restarting shipment: %v %v ...\n", shipmentName, shipment.Env)
		shipmentEnv := shipment.Env
		Restart(username, token, shipmentName, shipmentEnv)
		fmt.Println("done")

	} //shipments
}
