package cmd

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

//BuildTokenOutput represents an object that can be written to stdout and formatted
type BuildTokenOutput struct {
	Shipment    string
	Environment string
	CiCdEnvVar  string
	Token       string
}

func init() {
	RootCmd.AddCommand(buildTokenCmd)
	buildTokenCmd.AddCommand(listBuildTokensCmd)
}

// buildTokenCmd represents the buildtoken command
var buildTokenCmd = &cobra.Command{
	Use:   "buildtoken",
	Short: "manage harbor build tokens",
	Long:  `manage harbor build tokens`,
	Run:   buildTokens,
}

// listBuildTokensCmd represents the buildtoken command
var listBuildTokensCmd = &cobra.Command{
	Use:     "list",
	Short:   "list harbor build tokens for shipments in harbor-compose.yml",
	Long:    `list harbor build tokens for shipments in harbor-compose.yml`,
	Run:     listBuildTokens,
	Aliases: []string{"ls"}}

func buildTokens(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func listBuildTokens(cmd *cobra.Command, args []string) {

	//ensure user is logged in
	username, token, err := Login()
	if err != nil {
		log.Fatal(err)
	}

	//read the harbor compose file
	harborCompose := DeserializeHarborCompose(HarborComposeFile)
	if harborCompose.Shipments == nil || len(harborCompose.Shipments) == 0 {
		fmt.Println("no shipments found")
		return
	}

	//print table header
	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.DiscardEmptyColumns)
	fmt.Println("")
	fmt.Fprintln(w, "SHIPMENT\tENVIRONMENT\tCICD_ENVAR\tTOKEN\t")

	//create a formatted template
	tmpl, err := template.New("shipment-token").Parse("{{.Shipment}}\t{{.Environment}}\t{{.CiCdEnvVar}}\t{{.Token}}")
	if err != nil {
		log.Fatal(err)
	}

	//iterate shipments
	for shipmentName, shipment := range harborCompose.Shipments {

		//fetch the current state
		shipmentObject := GetShipmentEnvironment(username, token, shipmentName, shipment.Env)

		//build an object to pass to the template
		output := BuildTokenOutput{
			Shipment:    shipmentName,
			Environment: shipment.Env,
			CiCdEnvVar:  getBuildTokenName(shipmentName, shipment.Env),
			Token:       shipmentObject.BuildToken,
		}

		//execute the template with the data
		err = tmpl.Execute(w, output)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(w)
	}

	//flush the writer
	w.Flush()
	fmt.Println("")
}

func getBuildTokenName(shipment string, environment string) string {
	//SHIPMENT_ENV_TOKEN
	return fmt.Sprintf("%v_%v_TOKEN", strings.Replace(strings.ToUpper(shipment), "-", "_", -1), strings.ToUpper(environment))
}

func getBuildTokenEnvVar(shipment string, environment string) string {

	//look for envvar for this shipment/environment that matches naming convention: SHIPMENT_ENV_TOKEN
	envvar := getBuildTokenName(shipment, environment)
	if Verbose {
		log.Printf("looking for environment variable named: %v\n", envvar)
	}
	buildTokenEnvVar := os.Getenv(envvar)

	//validate build token
	if len(buildTokenEnvVar) == 0 {
		log.Fatalf("A shipment/environment build token is required. Please specify an environment variable named, %v", envvar)
	}

	return buildTokenEnvVar
}
