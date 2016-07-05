package cmd

import (
		"fmt"
		"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "View output from containers",
		Long:  ``,
		Run: logs,
}

func init() {
		RootCmd.AddCommand(logsCmd)
}

func logs(cmd *cobra.Command, args []string) {
		//read the harbor compose file
		var harborCompose = DeserializeHarborCompose(HarborComposeFile)
		//iterate shipments
		for shipmentName, shipment := range harborCompose.Shipments {
			fmt.Println("Logs For:  " + shipmentName + " " + shipment.Env)
			GetLogs(shipment.Barge, shipmentName, shipment.Env)
		}
}
