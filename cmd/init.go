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
	Short: "Interactively create a docker-compose.yml and a harbor-compose.yml file",
	Long: `This will ask you a bunch of questions, and then write a docker-compose.yml harbor-compose.yml for you.

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
	container := "mss-harbor-compose-app"
	tag := "0.1.0"
	publicPort := "80"
	internalPort := "5000"
	healthCheck := "/health"

	//harbor-compose
	name := "mss-harbor-compose-app"
	env := "dev"
	barge := "digital-sandbox"
	replicas := "2"
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
			registry = promptAndGetResponse("docker registry: (e.g., quay.io/turner) ")
			container = promptAndGetResponse("docker container name: (e.g., mss-harbor-compose-app) ")
			tag = promptAndGetResponse("version tag: (e.g., 0.1.0) ")
			publicPort = promptAndGetResponse("public port: (e.g., 80) ")
			internalPort = promptAndGetResponse("internal port: (e.g., 5000) ")
			healthCheck = promptAndGetResponse("health check: (e.g., /health) ")
		}

		//create a DockerCompose object
		dockerCompose = DockerCompose{
			Version:  "2",
			Services: map[string]*DockerComposeService{},
		}
		service := DockerComposeService{
			Build: ".",
			Image: fmt.Sprintf("%v/%v:%v", registry, container, tag),
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
		name = promptAndGetResponse("shipment name: (e.g., mss-my-app) ")
		env = promptAndGetResponse("shipment environment: (dev, qa, prod, etc.) ")
		barge = promptAndGetResponse("barge: (digital-sandbox, ent-prod, corp-sandbox, corp-prod, news, nba) ")
		replicas = promptAndGetResponse("replicas (how many container instances): ")
		intReplicas, err = strconv.Atoi(replicas)
		if err != nil {
			log.Fatalln("replicas must be a number")
		}
		group = promptAndGetResponse("group (mss, cnn, nba, ams, etc.): ")
		enableMonitoring = promptAndGetResponse("enableMonitoring (true|false): ")
		monitoring, err = strconv.ParseBool(enableMonitoring)
		if err != nil {
			check(errors.New("please enter true or false for enableMonitoring"))
		}
		hcTimeout = promptAndGetResponse("healthcheckTimeoutSeconds (1): ")
		healthcheckTimeoutSeconds, err = strconv.Atoi(hcTimeout)
		if err != nil {
			check(errors.New("please enter a valid number for healthcheckTimeoutSeconds"))
		}
		hcInterval = promptAndGetResponse("healthcheckIntervalSeconds (10): ")
		healthcheckIntervalSeconds, err = strconv.Atoi(hcInterval)
		if err != nil {
			check(errors.New("please enter a valid number for healthcheckIntervalSeconds"))
		}
		//healthcheckIntervalSeconds must be >= healthcheckTimeoutSeconds
		if !(healthcheckIntervalSeconds >= healthcheckTimeoutSeconds) {
			check(errors.New("healthcheckIntervalSeconds must be >= healthcheckTimeoutSeconds"))
		}
		property = promptAndGetResponse("property (turner.com, cnn.com, etc.): ")
		project = promptAndGetResponse("project: ")
		product = promptAndGetResponse("product: ")
		enableMonitoring = promptAndGetResponse("enableMonitoring (true|false): ")
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

	//look for existing docker-compose.yml to get containers
	if _, err := os.Stat(DockerComposeFile); err == nil {

		//parse the docker compose yaml to get the list of containers
		dockerCompose := DeserializeDockerCompose(DockerComposeFile)
		proj := dockerCompose.(*dockerProject.Project)

		//add all docker services as containers
		for _, service := range proj.ServiceConfigs.Keys() {
			composeShipment.Containers = append(composeShipment.Containers, service)
		}
	}

	//add single shipment to list
	harborCompose.Shipments[name] = composeShipment

	//if harbor-compose.yml exists, ask to overwrite
	write = false
	if _, err := os.Stat(HarborComposeFile); err == nil {
		fmt.Print("harbor-compose.yml already exists. Overwrite? ")
		write = askForConfirmation()
	} else { //not exists
		write = true
	}

	if write {
		SerializeHarborCompose(harborCompose, HarborComposeFile)
		fmt.Println("remember to docker build/push your image before running the up command")
	}

	fmt.Println("done")
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
