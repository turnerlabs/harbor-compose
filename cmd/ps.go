package cmd

import (
	"fmt"
	"html/template"
	"os"
	"strconv"
	"text/tabwriter"

	"log"

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

	//read the harbor compose file
	harborCompose := DeserializeHarborCompose(HarborComposeFile)

	//iterate Shipments
	for shipmentName, shipment := range harborCompose.Shipments {
		printShipmentStatus(shipmentName, shipment)
	}
}

func printShipmentStatus(name string, shipment ComposeShipment) {

	//fetch container status using helmit api
	shipmentStatus := GetShipmentStatus(shipment.Barge, name, shipment.Env)

	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.DiscardEmptyColumns)

	//create a formatted template
	tmpl, err := template.New("shipment").Parse("SHIPMENT\t{{.Shipment}}\t\nENVIRONMENT\t{{.Environment}}\t\nSTATUS\t{{.Status}}\t\nCONTAINERS\t{{.Containers}}\t\nREPLICAS\t{{.Replicas}}\t")

	fmt.Fprintln(w)

	if err != nil {
		log.Fatal(err)
	}

	//build an object to pass to the template
	shipmentOutput := ShipmentStatusOutput{
		Shipment:    name,
		Environment: shipment.Env,
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

	fmt.Fprintln(w, "IMAGE\tSTARTED\tSTATUS\tRESTARTS\t")

	for _, container := range shipmentStatus.Status.Containers {

		started := container.State.Running.StartedAt.String()
		if started != "" {
			started = humanize.Time(container.State.Running.StartedAt)
		}

		//create an object representing data
		output := ContainerStatusOutput{
			Image:    container.Image,
			Started:  started,
			Status:   container.Status,
			Restarts: strconv.Itoa(container.Restarts),
		}

		//create a formatted template
		tmpl, err := template.New("test").Parse("{{.Image}}\t{{.Started}}\t{{.Status}}\t{{.Restarts}}\t")

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
