package cmd

import (
		"fmt"
		"strings"
	  "time"
		"sort"
		"encoding/json"
		"github.com/spf13/cobra"
)

var Time bool
var Separate bool
// logsCmd represents the logs command
var logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "View output from containers",
		Long:  ``,
		Run: logs,
}

func init() {
	  logsCmd.PersistentFlags().BoolVarP(&Time, "time", "t", false, "append time to logs")
		logsCmd.PersistentFlags().BoolVarP(&Separate, "separate", "s", false, "print logs by each container")
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

			if Separate == true {
			    printSeparateLogs(helmitObject)
			} else {
          printMergedLogs(helmitObject)
			}
		}
}

// Object that contains a containers logs
type LogsObject struct {
	Name string
	Id string
	Image string
	Logs Logs
}

type LogObject struct {
	Time time.Time
	Log string
}

type Logs []LogObject

func (slice Logs) Len() int {
	return len(slice)
}

func (slice Logs) Less(i, j int) bool {
	return slice[i].Time.Before(slice[j].Time)
}

func (slice Logs) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func printMergedLogs(shipment HelmitResponse) {
	layout := "2006-01-02T15:04:05.999999999Z"
	shipmentLogs := make([]LogsObject, len(shipment.Replicas))
	for _, provider := range shipment.Replicas {
			for _, container := range provider.Containers {
				  var containerLogs = Logs{}
					for _, log := range container.Logs {
							line := strings.Fields(log)
							timeValue, err := time.Parse(layout, line[0])
							if err != nil {
									timeValue, err = time.Parse(layout, line[0][:1])
									if err != nil {
											continue
									}
							}

							var logObject = LogObject{}

							logObject.Time = timeValue
							logObject.Log = strings.Join(line, " ")

							containerLogs = append(containerLogs, logObject)
					}
					var logsObject  = LogsObject{}
					logsObject.Name = container.Name
					logsObject.Id = container.Id
					logsObject.Image = container.Image
					logsObject.Logs = containerLogs
					shipmentLogs = append(shipmentLogs, logsObject)
			}
	}

	var mergedLogs Logs
	for _, logObject := range shipmentLogs {
		for _, logObj := range logObject.Logs {
			  newLog := logObject.Name + ":" + logObject.Id[0:5] + "  | "
				if Time == true {
				     newLog = newLog + logObj.Time.String()  + ", "
				}

				logObj.Log = newLog + logObj.Log + "\n"
			  mergedLogs = append(mergedLogs, logObj)
		}
	}

	sort.Sort(mergedLogs)

	for _, log := range mergedLogs {
			fmt.Printf(log.Log)
	}
}

// printShipmentLogs
// prints the logs separatly for each shipment
func printSeparateLogs(shipment HelmitResponse) {
	for _, provider := range shipment.Replicas {
			for _, container := range provider.Containers {
					fmt.Printf("--- Name: %s\n", container.Name)
					fmt.Printf("--- Id: %s\n", container.Id)
					fmt.Printf("--- Image %s\n", container.Image)
					for _, log := range container.Logs {
							line := strings.Fields(log)

							if Time == false {
								line = append(line[:0], line[1:]...)
							}

							fmt.Println(strings.Join(line, " "))
					}
			}
	}
}
