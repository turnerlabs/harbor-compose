package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var logTime bool
var separateLogs bool
var tail bool

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [logid ...]",
	Short: "View output from containers",
	Long:  "View output of containers. There are few options available to make this easier to view.",
	Example: `harbor-compose logs

Print logs for specified shipment environment
Examples:
harbor-compose logs --shipment my-shipment --environment dev
harbor-compose logs --s my-shipment --e dev

Stream the logs from all of your replicas
Examples:
harbor-compose logs --tail
harbor-compose logs -t

Print the logs by each container
Examples:
harbor-compose logs --separate
harbor-compose logs -S -T

You can also pass in arguments to the function to allow for n number of specific queries.
Examples:
harbor-compose logs 9e70dc6 14ffbf5`,
	Run:    logs,
	PreRun: preRunHook,
}

var logsShipment string
var logsEnvironment string

func init() {
	logsCmd.PersistentFlags().StringVarP(&logsShipment, "shipment", "s", "", "shipment name")
	logsCmd.PersistentFlags().StringVarP(&logsEnvironment, "environment", "e", "", "environment name")
	logsCmd.PersistentFlags().BoolVarP(&logTime, "time", "T", false, "append time to logs")
	logsCmd.PersistentFlags().BoolVarP(&separateLogs, "separate", "S", false, "print logs by each container")
	logsCmd.PersistentFlags().BoolVarP(&tail, "tail", "t", false, "continue to stream log output to stdout.")
	RootCmd.AddCommand(logsCmd)
}

// logs cli command
// Usage: run inside folder with harbor-compose.yml file
// Flags: -t: adds time to the logs
// TODO: add the rest of the flags to match docker-compose
func logs(cmd *cobra.Command, args []string) {

	//make sure user is authenticated
	username, token, err := Login()
	check(err)

	//determine which shipment/environments user wants logs for
	inputShipmentEnvironments := getShipmentEnvironmentsFromInput(logsShipment, logsEnvironment)

	//iterate shipment/environments
	for _, t := range inputShipmentEnvironments {
		shipment := t.Item1
		env := t.Item2

		//lookup the shipment environment
		shipmentEnvironment := GetShipmentEnvironment(username, token, shipment, env)

		//lookup the provider
		provider := ec2Provider(shipmentEnvironment.Providers)

		fmt.Printf("Logs For:  %s %s\n", shipment, env)
		if len(args) > 0 && Verbose == true {
			fmt.Println("Make sure the ID is either the 7 char shortstring of the container or the entire ID")
			for _, arg := range args {
				fmt.Printf("Getting Logs for Container:  %s\n", arg)
			}
		}

		helmitObject := HelmitResponse{}

		var response = GetLogs(provider.Barge, shipment, env)
		err := json.Unmarshal([]byte(response), &helmitObject)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(args)

		if separateLogs {
			printSeparateLogs(helmitObject, args)
		} else {
			printMergedLogs(helmitObject, args)
		}
	}
}

// logsObject that contains a containers logs
type logsObject struct {
	Name      string
	ID        string
	Image     string
	Logstream string
	Logs      Logs
}

// logObject is a log object
type logObject struct {
	Time time.Time
	Log  string
}

// Logs is a list
type Logs []logObject

func (slice Logs) Len() int {
	return len(slice)
}

func (slice Logs) Less(i, j int) bool {
	return slice[i].Time.Before(slice[j].Time)
}

func (slice Logs) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// check to see if string is in the array
// http://stackoverflow.com/questions/15323767/does-golang-have-if-x-in-construct-similar-to-python
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// parseContainerLog will parse a log from docker and create an object containing needed information
func parseContainerLog(log string) (logObj logObject, errstring string) {
	layout := time.RFC3339
	line := strings.Fields(log)
	errstring = ""
	if len(line) > 2 {
		timeValue, err := time.Parse(layout, line[0])
		if err != nil {
			errstring = "Could not Parse"
		}

		logObj.Time = timeValue
		line = append(line[:0], line[1:]...)
		logObj.Log = strings.Join(line, " ")
	} else {
		errstring = "Format was off."
	}

	return
}

func printMergedLogs(shipment HelmitResponse, ids []string) {
	shipmentLogs := []logsObject{}
	for _, provider := range shipment.Replicas {
		for _, container := range provider.Containers {

			if len(ids) > 0 && !stringInSlice(container.ID, ids) && !stringInSlice(container.ID[0:7], ids) {
				continue
			}

			var containerLogs = Logs{}
			for _, logstring := range container.Logs {
				parsedLog, err := parseContainerLog(logstring)
				if err != "" {
					continue
				}
				containerLogs = append(containerLogs, parsedLog)
			}

			// set current log object
			var logsObject = logsObject{}
			logsObject.Name = container.Name
			logsObject.ID = container.ID
			logsObject.Image = container.Image
			logsObject.Logs = containerLogs
			logsObject.Logstream = container.Logstream
			shipmentLogs = append(shipmentLogs, logsObject)

		}
	}

	var mergedLogs Logs
	for _, logObject := range shipmentLogs {
		for _, logObj := range logObject.Logs {
			newLog := logObject.Name + ":" + logObject.ID[0:7] + "  | "
			if logTime == true {
				newLog = newLog + logObj.Time.String() + ", "
			}

			logObj.Log = newLog + logObj.Log + "\n"
			mergedLogs = append(mergedLogs, logObj)
		}
	}

	sort.Sort(mergedLogs)

	for _, log := range mergedLogs {
		fmt.Printf(log.Log)
	}

	if tail == true {
		for _, streamObj := range shipmentLogs {
			go followStream(streamObj)
		}
		var input string
		fmt.Scanln(&input)
	}
}

// followStream will take a logsObject param and print out all data that comes from the
// logStream field. This is a normal http response, which never ends.
func followStream(streamObj logsObject) {
	stream := strings.Replace(streamObj.Logstream, "tail=500", "tail=0", -1)
	streamer, err := GetLogStreamer(stream)
	check(err)
	for {
		line, streamErr := streamer.ReadBytes('\n')
		check(streamErr)
		logObj, err := parseContainerLog(string(line)[8:])

		if err != "" {
			fmt.Println(err)
			fmt.Println(string(line))
			continue
		}

		newLog := streamObj.Name + ":" + streamObj.ID[0:7] + "  | "
		if logTime == true {
			newLog = newLog + logObj.Time.String() + ", "
		}

		logObj.Log = newLog + logObj.Log
		fmt.Println(logObj.Log)
	}
}

// printShipmentLogs
// prints the logs separatly for each shipment
func printSeparateLogs(shipment HelmitResponse, ids []string) {
	for _, provider := range shipment.Replicas {
		for _, container := range provider.Containers {

			if len(ids) > 0 && !stringInSlice(container.ID, ids) && !stringInSlice(container.ID[0:7], ids) {
				continue
			}

			fmt.Printf("--- Name: %s\n", container.Name)
			fmt.Printf("--- Id: %s\n", container.ID)
			fmt.Printf("--- Image %s\n", container.Image)
			for _, log := range container.Logs {
				line := strings.Fields(log)

				if len(line) > 2 && logTime == false {
					line = append(line[:0], line[1:]...)
				}

				fmt.Println(strings.Join(line, " "))
			}

			if tail == true {
				logsObj := logsObject{}
				logsObj.Logstream = container.Logstream
				logsObj.Name = container.Name
				logsObj.ID = container.ID
				go followStream(logsObj)
			}
		}
	}

	if tail == true {
		var input string
		fmt.Scanln(&input)
	}
}
