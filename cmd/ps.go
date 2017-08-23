package cmd

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"log"

	"github.com/docker/libcompose/project"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "Lists shipment and container status",
	Long:  ``,
	Run:   ps,
}

func init() {
	RootCmd.AddCommand(psCmd)
}

func ps(cmd *cobra.Command, args []string) {

	//read the compose files
	dockerCompose, harborCompose := unmarshalComposeFiles(DockerComposeFile, HarborComposeFile)

	doPs(dockerCompose, harborCompose)
}

func doPs(dockerCompose project.APIProject, harborCompose HarborCompose) {

	//iterate Shipments
	for shipmentName, shipment := range harborCompose.Shipments {

		//fetch container status using helmit api
		shipmentStatus := GetShipmentStatus(shipment.Barge, shipmentName, shipment.Env)

		//get shipment's primary port
		primaryPort, err := getShipmentPrimaryPort(dockerCompose, shipment)
		if err != nil {
			log.Fatal(err)
		}

		//get shipment endpoint
		endpoint := getShipmentEndpoint(shipmentName, shipment.Env, "ec2", primaryPort)

		//print status to console
		printShipmentStatus(shipmentName, shipment, shipmentStatus, endpoint)
	}
}

func printShipmentStatus(name string, shipment ComposeShipment, shipmentStatus *ShipmentStatus, endpoint string) {

	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.DiscardEmptyColumns)

	//create a formatted template
	tmpl, err := template.New("shipment").Parse("SHIPMENT:\t{{.Shipment}}\t\nENVIRONMENT:\t{{.Environment}}\t\nENDPOINT:\t{{.Endpoint}}\t\nSTATUS:\t{{.Status}}\t\nCONTAINERS:\t{{.Containers}}\t\nREPLICAS:\t{{.Replicas}}\t")

	fmt.Fprintln(w)

	if err != nil {
		log.Fatal(err)
	}

	//build an object to pass to the template
	shipmentOutput := ShipmentStatusOutput{
		Shipment:    name,
		Environment: shipment.Env,
		Endpoint:    endpoint,
		Status:      shipmentStatus.Status.Phase,
		Containers:  strconv.Itoa(len(shipment.Containers)),
		Replicas:    strconv.Itoa(shipment.Replicas),
	}

	//execute the template with the data
	err = tmpl.Execute(w, shipmentOutput)
	if err != nil {
		log.Fatal(err)
	}
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

		if err != nil {
			log.Fatal(err)
		}

		//execute the template with the data
		err = tmpl.Execute(w, output)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintln(w)
	}
	w.Flush()

	fmt.Println("-----")
}

//returns the primary port of a shipment
func getShipmentPrimaryPort(dockerCompose project.APIProject, shipment ComposeShipment) (string, error) {

	//get primary port of 1st service
	for _, container := range shipment.Containers {
		serviceConfig, success := dockerCompose.GetServiceConfig(container)
		if !success {
			return "", errors.New("error getting service config")
		}

		if serviceConfig.Ports == nil {
			return "", errors.New("no ports found")
		}

		parsedPort := strings.Split(serviceConfig.Ports[0], ":")
		return parsedPort[0], nil
	}
	return "", errors.New("no shipment port found")
}

//returns a shipment endpoint
func getShipmentEndpoint(shipment string, environment string, provider string, port string) string {
	//80 -> http://xxx
	//443 -> https://xxx
	//? -> http://xxx:?
	protocol := "http"
	portString := ""
	if port == "443" {
		protocol = "https"
	}
	if port != "80" && port != "443" {
		portString = ":" + port
	}
	return fmt.Sprintf("%s://%s.%s.services.%s.dmtio.net%s", protocol, shipment, environment, provider, portString)
}
