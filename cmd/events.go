package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"

	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// eventsCmd represents the events command
var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Show container orchestration events",
	Long:  "Show container orchestration events for the shipment environments in your yaml files or specified via command line arguments",
	Example: `harbor-compose events
harbor-compose events --type all

# show only normal events
harbor-compose events --type normal

# show only warning events
harbor-compose events --type warning

# show full event messages
hargor-compose events -m

# show events for a particular shipment environment
harbor-compose events --shipment my-shipment --environment dev
harbor-compose events -s my-shipment -e dev
`,
	Run:    runEvents,
	PreRun: preRunHook,
}

var eventsShipment string
var eventsEnvironment string
var eventsType string
var eventsFullMessage bool

func init() {
	eventsCmd.PersistentFlags().StringVarP(&eventsShipment, "shipment", "s", "", "shipment name")
	eventsCmd.PersistentFlags().StringVarP(&eventsEnvironment, "environment", "e", "", "environment name")
	eventsCmd.PersistentFlags().StringVarP(&eventsType, "type", "t", "all", "specify what level of events you would like to see (normal, warning, or all)")
	eventsCmd.PersistentFlags().BoolVarP(&eventsFullMessage, "message", "m", false, "include the full message")
	RootCmd.AddCommand(eventsCmd)
}

// events your shipment
func runEvents(cmd *cobra.Command, args []string) {

	//make sure user is authenticated
	username, token, err := Login()
	check(err)

	//determine which shipment/environments user wants status for
	inputShipmentEnvironments, _ := getShipmentEnvironmentsFromInput(eventsShipment, eventsEnvironment)

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

		//fetch events from helmit
		events := GetShipmentEvents(provider.Barge, shipment, env)

		//sort by LastTimestamp (the last time this event happened)
		sort.Slice(events.Events, func(i, j int) bool {
			return events.Events[i].LastTimestamp.After(events.Events[j].LastTimestamp)
		})

		//render
		if len(events.Events) > 0 {

			if eventsFullMessage {
				printShipmentEventMessages(events)
			} else {
				printShipmentEvents(events)
			}
			fmt.Println("-----")

		} else {
			fmt.Println("no events found")
			fmt.Println()
			fmt.Println("note that events only occur around a deployment or an unhealthy application and eventually disappear when an app becomes healthy")
		}
	}
}

func printShipmentEvents(events *ShipmentEventResult) {
	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.DiscardEmptyColumns)

	fmt.Println()
	fmt.Fprintln(w, "TYPE\tREASON\tMESSAGE\tTIME\tCOUNT")

	//create a formatted template
	tmpl, err := template.New("events").Parse("{{.Type}}\t{{.Reason}}\t{{.Message}}\t{{.StartTime}}\t{{.Count}}\t")
	check(err)

	for _, event := range events.Events {

		//filter events for specified type
		if strings.ToLower(eventsType) != "all" && strings.ToLower(eventsType) != strings.ToLower(event.Type) {
			continue
		}

		event.StartTime = humanize.Time(event.LastTimestamp)

		//truncate message
		if !eventsFullMessage {
			truncatedLength := 60
			if len(event.Message) > truncatedLength {
				event.Message = event.Message[:truncatedLength]
			}
		}

		//execute the template with the data
		err = tmpl.Execute(w, event)
		check(err)
		fmt.Fprintln(w)
	}
	w.Flush()
}

func printShipmentEventMessages(events *ShipmentEventResult) {
	fmt.Println()
	for _, event := range events.Events {

		//filter events for specified type
		if strings.ToLower(eventsType) != "all" && strings.ToLower(eventsType) != strings.ToLower(event.Type) {
			continue
		}

		fmt.Printf("%s - %s\n", event.Reason, event.Message)
		fmt.Println()
	}
}
