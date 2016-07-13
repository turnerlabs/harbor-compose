package cmd

import (
		"fmt"
		"strings"
		"encoding/json"
		"github.com/spf13/cobra"
)

var Time bool
// logsCmd represents the logs command
var logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "View output from containers",
		Long:  ``,
		Run: logs,
}

func init() {
	  RootCmd.PersistentFlags().BoolVarP(&Time, "time", "t", false, "append time to logs")
		RootCmd.AddCommand(logsCmd)
}

// logs cli command
// Usage: run inside folder with harbor-compose.yml file
// Flags: -t: adds time to the logs
// TODO: add the rest of the flags to match docker-compose
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
					for _, container := range provider.Containers {
						  fmt.Printf("--- Name: %s\n", container.Name)
							fmt.Printf("--- Id: %s\n", container.Id)
							fmt.Printf("--- Image %s\n", container.Image)
							for _, log := range container.Logs {
								  line := strings.Split(log, ",")
									line = strings.Split(line[0], " ")

									if Time == false {
									  line = append(line[:0], line[1:]...)
									}

                  fmt.Println(strings.Join(line, " "))
							}
					}
			}
		}
}
