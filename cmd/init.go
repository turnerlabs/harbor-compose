package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	dockerProject "github.com/docker/libcompose/project"
	"github.com/spf13/cobra"
)

var yesUseDefaults bool

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactively create main.tf, docker-compose.yml, and harbor-compose.yml files",
	Long: `This will ask you a bunch of questions, and then write a main.tf, docker-compose.yml, and harbor-compose.yml for you.

If you invoke it with -y or --yes it will use only defaults and not prompt you for any options.`,
	Run: initHarborCompose,
}

func init() {
	initCmd.PersistentFlags().BoolVarP(&yesUseDefaults, "yes", "y", false, "don't ask questions and use defaults")
	RootCmd.AddCommand(initCmd)
}

func initHarborCompose(cmd *cobra.Command, args []string) {

	//write a gitignored/dockerignored hidden.env as a placeholder for users to add secrets
	//if it already exists, prompt to overwrite
	//note that this file needs to be written before reference is added/parsed in docker-compose.yml
	write := false
	if _, err := os.Stat(hiddenEnvFileName); err == nil {
		fmt.Print(hiddenEnvFileName + " already exists. Overwrite? ")
		write = askForConfirmation()
	} else { //not exists
		write = true
	}
	if write {
		sampleContents := "#FOO=bar\n"
		err := ioutil.WriteFile(hiddenEnvFileName, []byte(sampleContents), 0644)
		check(err)
		sensitiveFiles := []string{hiddenEnvFileName}
		appendToFile(".gitignore", sensitiveFiles)
		appendToFile(".dockerignore", sensitiveFiles)
	}

	//docker-compose
	registry := "quay.io/turner"
	container := "mss-harbor-app"
	tag := "0.1.0"
	publicPort := "80"
	internalPort := "5000"
	healthCheck := "/health"

	//harbor-compose
	name := "mss-harbor-app"
	env := "dev"
	barge := "digital-sandbox"
	replicas := "4"
	group := "mss"
	property := "turner"
	project := "turner"
	product := "turner"
	enableMonitoring := "true"
	hcTimeout := "1"
	hcInterval := "10"

	//if docker-compose.yml doesn't exist, then create one
	var dockerCompose DockerCompose
	if _, err := os.Stat(DockerComposeFile); err != nil {

		if !yesUseDefaults {
			registry = promptAndGetResponse("docker registry: (e.g., quay.io/turner) ", registry)
			container = promptAndGetResponse("docker container name: (e.g., mss-harbor-app) ", container)
			tag = promptAndGetResponse("version tag: (e.g., 0.1.0) ", tag)
			publicPort = promptAndGetResponse("public port: (e.g., 80) ", publicPort)
			internalPort = promptAndGetResponse("internal port: (e.g., 5000) ", internalPort)
			healthCheck = promptAndGetResponse("health check: (e.g., /health) ", healthCheck)
		}

		//create a DockerCompose object
		dockerCompose = DockerCompose{
			Version:  "2",
			Services: map[string]*DockerComposeService{},
		}

		image := fmt.Sprintf("%v/%v:%v", registry, container, tag)

		//substitute default with default backend
		if image == "quay.io/turner/mss-harbor-app:0.1.0" {
			image = "quay.io/turner/turner-defaultbackend:0.2.0"
		}

		service := DockerComposeService{
			Build: ".",
			Image: image,
			Ports: []string{fmt.Sprintf("%v:%v", publicPort, internalPort)},
			Environment: map[string]string{
				"HEALTHCHECK": healthCheck,
			},
			EnvFile: []string{hiddenEnvFileName},
		}
		dockerCompose.Services[container] = &service

		//write docker-compose.yml
		SerializeDockerCompose(dockerCompose, DockerComposeFile)
	}

	//ask questions for harbor-compose.yml
	var (
		intReplicas                int
		monitoring                 bool
		healthcheckTimeoutSeconds  int
		healthcheckIntervalSeconds int
		err                        error
	)
	if !yesUseDefaults {
		name = promptAndGetResponse("shipment name: (e.g., mss-harbor-app) ", name)
		env = promptAndGetResponse("shipment environment: (dev, qa, prod, etc.) ", env)
		barge = promptAndGetResponse("barge: (digital-sandbox, ent-prod, corp-sandbox, corp-prod, news, nba) ", barge)
		replicas = promptAndGetResponse("how many container instances: (e.g., 4) ", replicas)
		intReplicas, err = strconv.Atoi(replicas)
		if err != nil {
			log.Fatalln("replicas must be a number")
		}
		group = promptAndGetResponse("group (mss, news, nba, ams, etc.): ", group)
		enableMonitoring = promptAndGetResponse("enable monitoring (true|false): ", enableMonitoring)
		monitoring, err = strconv.ParseBool(enableMonitoring)
		if err != nil {
			check(errors.New("please enter true or false for enableMonitoring"))
		}
		hcTimeout = promptAndGetResponse("healthcheck timeout seconds (1): ", hcTimeout)
		healthcheckTimeoutSeconds, err = strconv.Atoi(hcTimeout)
		if err != nil {
			check(errors.New("please enter a valid number for healthcheckTimeoutSeconds"))
		}
		hcInterval = promptAndGetResponse("healthcheck interval seconds (10): ", hcInterval)
		healthcheckIntervalSeconds, err = strconv.Atoi(hcInterval)
		if err != nil {
			check(errors.New("please enter a valid number for healthcheckIntervalSeconds"))
		}
		//healthcheckIntervalSeconds must be >= healthcheckTimeoutSeconds
		if !(healthcheckIntervalSeconds >= healthcheckTimeoutSeconds) {
			check(errors.New("healthcheckIntervalSeconds must be >= healthcheckTimeoutSeconds"))
		}
		property = promptAndGetResponse("property (turner.com, cnn.com, etc.): ", property)
		project = promptAndGetResponse("project: ", project)
		product = promptAndGetResponse("product: ", product)
	}

	//create a harbor compose object
	harborCompose := HarborCompose{
		Shipments: make(map[string]ComposeShipment),
	}

	monitoring, err = strconv.ParseBool(enableMonitoring)
	if err != nil {
		check(errors.New("please enter true or false for enableMonitoring"))
	}

	composeShipment := ComposeShipment{
		Env:                        env,
		Group:                      group,
		Property:                   property,
		Project:                    project,
		Product:                    product,
		EnableMonitoring:           &monitoring,
		HealthcheckTimeoutSeconds:  &healthcheckTimeoutSeconds,
		HealthcheckIntervalSeconds: &healthcheckIntervalSeconds,
	}

	composeShipment.Barge = barge
	composeShipment.Replicas = intReplicas

	//use existing docker-compose.yml to get containers
	dockerComposeProj := DeserializeDockerCompose(DockerComposeFile)
	proj := dockerComposeProj.(*dockerProject.Project)

	//add all docker services as containers
	for _, service := range proj.ServiceConfigs.Keys() {
		composeShipment.Containers = append(composeShipment.Containers, service)
	}

	//add single shipment to list
	harborCompose.Shipments[name] = composeShipment

	//if harbor-compose.yml exists, ask to overwrite
	write = true
	if _, err := os.Stat(HarborComposeFile); err == nil {
		fmt.Print("harbor-compose.yml already exists. Overwrite? ")
		write = askForConfirmation()
	}
	if write {
		SerializeHarborCompose(harborCompose, HarborComposeFile)
	}

	//transform compose yaml into a ShipmentEnvironment object
	shipmentEnvironment := transformComposeToShipmentEnvironment(name, composeShipment, dockerComposeProj)

	//generate a main.tf and write it to disk
	generateAndWriteTerraformSource(&shipmentEnvironment, &harborCompose, false)

	fmt.Println("done")
	fmt.Println()
	fmt.Println("use terraform plan/apply to manage your infrastructure")
	fmt.Println("use docker-compose build/push, harbor-compose up/deploy to manage your application")
}

func promptAndGetResponse(question string, defaultResponse string) string {
	fmt.Print(question)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil && err.Error() != "unexpected newline" {
		log.Fatal(err)
	}
	if response == "" {
		response = defaultResponse
	}
	return response
}
