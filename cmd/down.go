package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var deleteShipmentEnvironment bool

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop your application",
	Long:  `The down command brings your application down and optionally deletes your shipment environment.`,
	Run:   down,
}

func init() {
	downCmd.PersistentFlags().BoolVarP(&deleteShipmentEnvironment, "delete", "d", false, "deletes your shipment environment")
	RootCmd.AddCommand(downCmd)
}

func down(cmd *cobra.Command, args []string) {

	//read the harbor compose file
	var harborCompose = DeserializeHarborCompose(HarborComposeFile)

	_, _, err := Login()
	if err != nil {
		log.Fatalf(err.Error())
	}

	//iterate shipments
	for shipmentName, shipment := range harborCompose.Shipments {
		fmt.Printf("Stopping %v ...\n", shipmentName)

		if Verbose {
			log.Println("processing  " + shipmentName + "/" + shipment.Env)
			log.Println(shipment.Containers)
		}

		//override replicas
		shipment.Replicas = 0

		//update shipment level configuration
		UpdateShipment(shipmentName, shipment)

		//trigger shipment
		Trigger(shipmentName, shipment.Env)

		if deleteShipmentEnvironment {
			fmt.Printf("Deleting %v %v ...\n", shipmentName, shipment.Env)
			DeleteShipmentEnvironment(shipmentName, shipment.Env)
		}

		fmt.Println("done")
	}
}
