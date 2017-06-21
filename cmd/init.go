package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var yesUseDefaults bool

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactively create a harbor-compose.yml file",
	Long: `This will ask you a bunch of questions, and then write a harbor-compose.yml for you.

If you invoke it with -y or --yes it will use only defaults and not prompt you for any options.`,
	Run: initHarborCompose,
}

func init() {
	initCmd.PersistentFlags().BoolVarP(&yesUseDefaults, "yes", "y", false, "don't ask questions and use defaults")
	RootCmd.AddCommand(initCmd)
}

func initHarborCompose(cmd *cobra.Command, args []string) {

	name := "mss-harbor-compose-app"
	env := "dev"
	barge := "digital-sandbox"
	replicas := "2"
	group := "mss"
	property := "turner.com"
	project := "poc"
	product := "poc"

	//ask questions
	if !yesUseDefaults {
		name = promptAndGetResponse("shipment name: (e.g., mss-my-app) ")
		env = promptAndGetResponse("shipment environment: (dev, qa, prod, etc.) ")
		barge = promptAndGetResponse("barge: (cnn, nba, digital-sandbox, ent-prod, corp-sandbox, corp-prod) ")
		replicas = promptAndGetResponse("replicas (how many container instances): ")
		group = promptAndGetResponse("group (mss, cnn, nba, ams, etc.): ")
		property = promptAndGetResponse("property (turner.com, cnn.com, etc.): ")
		project = promptAndGetResponse("project: ")
		product = promptAndGetResponse("product: ")
	}

	//create a harbor compose object
	harborCompose := HarborCompose{
		Shipments: make(map[string]ComposeShipment),
	}

	composeShipment := ComposeShipment{
		Env:      env,
		Group:    group,
		Property: property,
		Project:  project,
		Product:  product,
	}

	composeShipment.Barge = barge
	intReplicas, err := strconv.Atoi(replicas)
	if err != nil {
		log.Fatalln("replicas must be a number")
	}
	composeShipment.Replicas = intReplicas

	//look for existing docker-compose.yml to get containers
	if _, err := os.Stat(DockerComposeFile); err == nil {

		//parse the docker compose yaml to get the list of containers
		dockerCompose, _ := DeserializeDockerCompose(DockerComposeFile)

		//add all docker services as containers
		for container := range dockerCompose.Services {
			composeShipment.Containers = append(composeShipment.Containers, container)
		}
	}

	//add single shipment to list
	harborCompose.Shipments[name] = composeShipment

	//if harbor-compose.yml exists, ask to overwrite
	write := false
	if _, err := os.Stat(HarborComposeFile); err == nil {
		fmt.Print("harbor-compose.yml already exists. Overwrite? ")
		write = askForConfirmation()
	} else { //not exists
		write = true
	}

	if write {
		SerializeHarborCompose(harborCompose, HarborComposeFile)
		fmt.Println("done")
	}

}

func promptAndGetResponse(question string) string {
	fmt.Print(question)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}
	return response
}
