package cmd

import (
		"fmt"
		"encoding/json"
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
			helmitObject := HelmitResponse{}
			var response = GetLogs(shipment.Barge, shipmentName, shipment.Env)
			err := json.Unmarshal([]byte(response), &helmitObject)
			if err != nil {
			    fmt.Println(err)
			}
			for _, provider := range helmitObject.Replicas {
					fmt.Println("--------------------------------------------------------------------------------------------------------------")
					fmt.Println("--------------------------------------- Host " + provider.Host)
					fmt.Println("--------------------------------------------------------------------------------------------------------------")
					for _, container := range provider.Containers {
					  	fmt.Println("--------------------------------------------------------------------------------------------------------------")
						  fmt.Println("--------------------------------------------------------------------------------------------------------------")
						  fmt.Println("--------------------------------------- Logs For " + container.Name)
					    fmt.Println("--------------------------------------------------------------------------------------------------------------")
							fmt.Println("--------------------------------------------------------------------------------------------------------------")
              fmt.Println(container.Logs)
					}
			}
		}
}
