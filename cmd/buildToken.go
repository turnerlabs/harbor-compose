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

var listEnvironmentOverride string

func init() {
	RootCmd.AddCommand(buildTokenCmd)

	buildTokenCmd.AddCommand(getBuildTokenCmd)

	listBuildTokensCmd.PersistentFlags().StringVarP(&listEnvironmentOverride, "env", "e", "", "list build tokens for an alternative environment.")
	buildTokenCmd.AddCommand(listBuildTokensCmd)
}

// buildTokenCmd represents the buildtoken command
var buildTokenCmd = &cobra.Command{
	Use:   "buildtoken",
	Short: "manage harbor build tokens",
	Long:  `manage harbor build tokens`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// listBuildTokensCmd represents the buildtoken command
var listBuildTokensCmd = &cobra.Command{
	Use:   "list",
	Short: "list harbor build tokens for shipments in harbor-compose.yml",
	Long:  `list harbor build tokens for shipments in harbor-compose.yml`,
	Example: `
harbor-compose buildtoken ls

SHIPMENT             ENVIRONMENT   CICD_ENVAR                     TOKEN
mss-poc-sqs-web      dev           MSS_POC_SQS_WEB_DEV_TOKEN      3xFVlltLZ7JwPH20Km75DrpMwOk2a4yq
mss-poc-sqs-worker   dev           MSS_POC_SQS_WORKER_DEV_TOKEN   2N3QFkQkdilwj34ezS2JTxwt6Fn3yuA8	

OR

harbor-compose buildtoken ls -e qa

SHIPMENT             ENVIRONMENT   CICD_ENVAR                     TOKEN
mss-poc-sqs-web      qa            MSS_POC_SQS_WEB_QA_TOKEN       ihtvPrAH84ULVm6IC7LjWvXUgEhr7cnQ
mss-poc-sqs-worker   qa            MSS_POC_SQS_WORKER_QA_TOKEN    Y3Jk0DmMaUsoWO8mbI2Edn9Ixhwj14Vd
	`,
	Run:     listBuildTokens,
	Aliases: []string{"ls"}}

var getBuildTokenCmd = &cobra.Command{
	Use:   "get",
	Short: "displays a build token for the requested shipment and environment",
	Long: `
displays a build token for the requested shipment and environment

will prompt for a shipment and environment and display the build token	
`,
	Example: `
harbor-compose buildtoken get

Shipment: mss-poc-sqs-worker
Environment: dev

SHIPMENT             ENVIRONMENT   CICD_ENVAR                     TOKEN
mss-poc-sqs-worker   dev           MSS_POC_SQS_WORKER_DEV_TOKEN   2N3QFkQkdilwj34ezS2JTxwt6Fn3yuA8

OR 

harbor-compose buildtoken get mss-poc-sqs-worker dev

SHIPMENT             ENVIRONMENT   CICD_ENVAR                     TOKEN
mss-poc-sqs-worker   dev           MSS_POC_SQS_WORKER_DEV_TOKEN   2N3QFkQkdilwj34ezS2JTxwt6Fn3yuA8
`,
	Run: getBuildToken,
}

func listBuildTokens(cmd *cobra.Command, args []string) {

	//ensure user is logged in
	username, authToken, err := Login()
	if err != nil {
		log.Fatal(err)
	}

	//read the harbor compose file
	harborCompose := DeserializeHarborCompose(HarborComposeFile)
	if harborCompose.Shipments == nil || len(harborCompose.Shipments) == 0 {
		fmt.Println("no shipments found")
		return
	}

	//do the work
	internalListBuildTokens(harborCompose.Shipments, username, authToken)
}

func internalListBuildTokens(shipmentMap map[string]ComposeShipment, username string, authToken string) {

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
	for shipmentName, shipment := range shipmentMap {

		//allow --env flag to override environment specified in compose file
		shipmentEnv := shipment.Env
		if len(listEnvironmentOverride) > 0 {
			shipmentEnv = listEnvironmentOverride
		}

		//fetch the current state
		shipmentObject := GetShipmentEnvironment(username, authToken, shipmentName, shipmentEnv)
		if shipmentObject == nil {
			continue
		}

		//build an object to pass to the template
		output := BuildTokenOutput{
			Shipment:    shipmentName,
			Environment: shipmentEnv,
			CiCdEnvVar:  getBuildTokenName(shipmentName, shipmentEnv),
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

func getBuildToken(cmd *cobra.Command, args []string) {

	//ensure user is logged in
	username, authToken, err := Login()
	if err != nil {
		log.Fatal(err)
	}

	var shipment string
	var env string
	if len(args) >= 2 {
		shipment = args[0]
		env = args[1]
	} else {
		//prompt for shipment/environment
		fmt.Print("Shipment: ")
		shipment = askForString()
		fmt.Print("Environment: ")
		env = askForString()
	}

	//create a shipment map so that we can leverage internalListBuildTokens
	shipments := map[string]ComposeShipment{}
	shipments[shipment] = ComposeShipment{Env: env}

	//do the work
	internalListBuildTokens(shipments, username, authToken)
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
