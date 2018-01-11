package cmd

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "Lists shipment and container status",
	Long:  "Lists shipment and container status for shipment environments listed in a harbor-compose.yml file or using the flags.",
	Example: `harbor-compose ps
harbor-compose ps --shipment my-shipment --environment dev
harbor-compose ps -s my-shipment -e dev`,
	Run:    ps,
	PreRun: preRunHook,
}

var psShipment string
var psEnvironment string

func init() {
	psCmd.PersistentFlags().StringVarP(&psShipment, "shipment", "s", "", "shipment name")
	psCmd.PersistentFlags().StringVarP(&psEnvironment, "environment", "e", "", "environment name")
	RootCmd.AddCommand(psCmd)
}

func ps(cmd *cobra.Command, args []string) {

	//make sure user is authenticated
	username, token, err := Login()
	check(err)

	//determine which shipment/environments user wants status for
	inputShipmentEnvironments, _ := getShipmentEnvironmentsFromInput(psShipment, psEnvironment)

	//iterate shipment/environments
	for _, t := range inputShipmentEnvironments {
		shipment := t.Item1
		env := t.Item2

		//lookup the shipment environment
		shipmentEnvironment := GetShipmentEnvironment(username, token, shipment, env)
		if shipmentEnvironment == nil {
			fmt.Println(messageShipmentEnvironmentNotFound)
			return
		}		

		//lookup the provider
		provider := ec2Provider(shipmentEnvironment.Providers)

		//fetch container status using helmit api
		shipmentStatus := GetShipmentStatus(provider.Barge, shipment, env)

		//get shipment's primary port
		primaryPort, err := getShipmentPrimaryPort(shipmentEnvironment)
		check(err)

		//get shipment endpoint
		endpoint := getShipmentEndpoint(shipment, env, provider.Name, primaryPort)

		//print status to console
		printShipmentStatus(shipment, shipmentEnvironment, provider, shipmentStatus, endpoint)
	}
}

func printShipmentStatus(name string, shipment *ShipmentEnvironment, provider *ProviderPayload, shipmentStatus *ShipmentStatus, endpoint string) {

	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.DiscardEmptyColumns)

	//create a formatted template
	tmpl, err := template.New("shipment").Parse("SHIPMENT:\t{{.Shipment}}\t\nENVIRONMENT:\t{{.Environment}}\nBARGE:\t{{.Barge}}\t\nENDPOINT:\t{{.Endpoint}}\t\nSTATUS:\t{{.Status}}\t\nCONTAINERS:\t{{.Containers}}\t\nREPLICAS:\t{{.Replicas}}\t")

	fmt.Fprintln(w)
	check(err)

	//build an object to pass to the template
	shipmentOutput := ShipmentStatusOutput{
		Shipment:    name,
		Environment: shipment.Name,
		Barge:       provider.Barge,
		Endpoint:    endpoint,
		Status:      shipmentStatus.Status.Phase,
		Containers:  strconv.Itoa(len(shipment.Containers)),
		Replicas:    strconv.Itoa(provider.Replicas),
	}

	//execute the template with the data
	err = tmpl.Execute(w, shipmentOutput)
	check(err)
	w.Flush()

	//format containers
	fmt.Println("")
	fmt.Println("")

	fmt.Fprintln(w, "ID\tIMAGE\tSTATUS\tSTARTED\tRESTARTS\tLAST STATE\t")

	for _, container := range shipmentStatus.Status.Containers {

		//get the container state info
		state := container.State[container.Status]

		//started at
		started := ""
		if container.Status == "running" {
			started = humanize.Time(state.StartedAt)
		}
		if container.Status == "waiting" {
			started = state.Reason
		}

		//last state
		lastState := ""
		if container.LastState["terminated"] != (ContainerLastState{}) {
			lastState = "terminated " + humanize.Time(container.LastState["terminated"].FinishedAt)
		}

		//create an object representing data
		output := ContainerStatusOutput{
			ID:        container.ID[0:7],
			Image:     container.Image,
			Status:    container.Status,
			Started:   started,
			Restarts:  strconv.Itoa(container.Restarts),
			LastState: lastState,
		}

		//create a formatted template
		tmpl, err := template.New("replicas").Parse("{{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Started}}\t{{.Restarts}}\t{{.LastState}}\t")
		check(err)

		//execute the template with the data
		err = tmpl.Execute(w, output)
		check(err)

		fmt.Fprintln(w)
	}
	w.Flush()

	fmt.Println("-----")
}

//returns the primary port of a shipment
func getShipmentPrimaryPort(shipmentEnvironment *ShipmentEnvironment) (PortPayload, error) {
	for _, container := range shipmentEnvironment.Containers {
		for _, port := range container.Ports {
			if port.Primary {
				return port, nil
			}
		}
	}
	return PortPayload{}, errors.New("no shipment port found")
}

func getShipmentEndpoint(shipment string, environment string, provider string, port PortPayload) string {
	return fmt.Sprintf("%s://%s.%s.services.%s.dmtio.net:%v", port.Protocol, shipment, environment, provider, port.PublicPort)
}
