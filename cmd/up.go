package cmd

import (
	"fmt"
	"log"

	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start your application",
	Long: `The up command applies changes from your docker/harbor compose files and brings your application up on Harbor.  The up command:

- Create Harbor shipment(s) if needed
- Update container and shipment/environment level environment variables in Harbor
- Update container images in Harbor
- Update replicas in Harbor
- Trigger your shipment(s) in Harbor
	`,
	Run: up,
}

func init() {
	RootCmd.AddCommand(upCmd)
}

func up(cmd *cobra.Command, args []string) {
	username, token, _ := Login()

	//read the harbor compose file
	harborCompose := DeserializeHarborCompose(HarborComposeFile)

	//read the docker compose file
	dockerCompose := DeserializeDockerCompose(DockerComposeFile)

	//iterate shipments
	for shipmentName, shipment := range harborCompose.Shipments {
		fmt.Printf("Starting %v ...\n", shipmentName)

		if Verbose {
			log.Printf("processing shipment: %v/%v", shipmentName, shipment.Env)
		}

		//fetch the current state
		shipmentObject := GetShipmentEnvironment(username, token, shipmentName, shipment.Env)

		//creating a shipment is a different workflow than updating
		//bulk create a shipment if it doesn't exist
		if shipmentObject == nil {
			if Verbose {
				log.Println("shipment environment not found")
			}
			createShipment(username, token, shipmentName, dockerCompose, shipment)

		} else {
			//make changes to harbor based on compose files
			updateShipment(username, token, shipmentObject, shipmentName, dockerCompose, shipment)

			//TODO: desired state reconciliation
		}

		fmt.Println("done")

	} //shipments
}

func createShipment(username string, token string, shipmentName string, dockerCompose DockerCompose, shipment ComposeShipment) {
	userName, token, _ := Login()

	//map a ComposeShipment object (based on compose files) into
	//a new NewShipmentEnvironment object

	if Verbose {
		log.Println("creating shipment environment")
	}

	//create object used to create a new shipment environment from scratch
	newShipment := NewShipmentEnvironment{
		Username: userName,
		Token:    token,
		Info: NewShipmentInfo{
			Name:  shipmentName,
			Group: shipment.Group,
		},
	}

	//add shipment-level env vars
	newShipment.Info.Vars = make([]EnvVarPayload, 0)
	newShipment.Info.Vars = append(newShipment.Info.Vars, envVar("CUSTOMER", shipment.Group))
	newShipment.Info.Vars = append(newShipment.Info.Vars, envVar("PROPERTY", shipment.Property))
	newShipment.Info.Vars = append(newShipment.Info.Vars, envVar("PROJECT", shipment.Project))
	newShipment.Info.Vars = append(newShipment.Info.Vars, envVar("PRODUCT", shipment.Product))

	//create environment
	newShipment.Environment = NewEnvironment{
		Name: shipment.Env,
		Vars: make([]EnvVarPayload, 0),
	}

	//add environment-level env vars
	for name, value := range shipment.Environment {
		newShipment.Environment.Vars = append(newShipment.Environment.Vars, envVar(name, value))
	}

	//containers
	//iterate defined containers and apply container level updates
	newShipment.Containers = make([]NewContainer, 0)
	for _, container := range shipment.Containers {

		if Verbose {
			log.Printf("processing container: %v", container)
		}

		//lookup the container in the list of services in the docker-compose file
		dockerService := dockerCompose.Services[container]

		if dockerService.Image == "" {
			log.Fatalln("'image' is required in docker compose file")
		}

		//parse image:tag and map to name/version
		parsedImage := strings.Split(dockerService.Image, ":")

		//newContainer := NewContainer{
		newContainer := NewContainer{
			Name:    container,
			Image:   dockerService.Image,
			Version: parsedImage[1],
			Vars:    make([]EnvVarPayload, 0),
			Ports:   make([]PortPayload, 0),
		}

		//container-level env vars
		for name, value := range dockerService.Environment {
			newContainer.Vars = append(newContainer.Vars, envVar(name, value))
		}

		//map the docker compose service ports to harbor ports
		if len(dockerService.Ports) == 0 {
			log.Fatalln("At least one port mapping is required in docker compose file.")
		}

		parsedPort := strings.Split(dockerService.Ports[0], ":")

		//validate health check
		healthCheck := dockerService.Environment["HEALTHCHECK"]
		if healthCheck == "" {
			log.Fatalln("A container-level 'HEALTHCHECK' environment variable is required")
		}

		//map first port in docker compose to default primary HTTP "PORT"

		external, err := strconv.Atoi(parsedPort[0])
		if err != nil {
			log.Fatalln("invalid port")
		}
		internal, err := strconv.Atoi(parsedPort[1])
		if err != nil {
			log.Fatalln("invalid port")
		}

		primaryPort := PortPayload{
			Name:        "PORT",
			Value:       internal,
			PublicPort:  external,
			Primary:     true,
			Protocol:    "http",
			External:    false,
			Healthcheck: healthCheck,
		}

		newContainer.Ports = append(newContainer.Ports, primaryPort)

		//TODO: once Container/Port construct is added to harbor-compose.yml,
		//they should override these defaults

		//add container to list
		newShipment.Containers = append(newShipment.Containers, newContainer)
	}

	if shipment.Barge == "" {
		log.Fatalln("barge is required for a shipment")
	}

	//add default ec2 provider
	provider := NewProvider{
		Name:     "ec2",
		Barge:    shipment.Barge,
		Replicas: shipment.Replicas,
		Vars:     make([]EnvVarPayload, 0),
	}

	//add provider
	newShipment.Providers = append(newShipment.Providers, provider)

	//push the new shipment/environment up to harbor
	SaveNewShipmentEnvironment(username, token, newShipment)

	//trigger shipment
	success, messages := Trigger(shipmentName, shipment.Env)

	for _, msg := range messages {
		fmt.Println(msg)
	}

	if success {
		fmt.Println("Please allow up to 5 minutes for DNS changes to take effect.")
	}
}

func updateShipment(username string, token string, currentShipment *ShipmentEnvironment, shipmentName string, dockerCompose DockerCompose, shipment ComposeShipment) {

	//map a ComposeShipment object (based on compose files) into
	//a series of API call to update a shipment

	//iterate defined containers and apply container level updates
	for _, container := range shipment.Containers {
		if Verbose {
			log.Printf("processing container: %v", container)
		}

		//lookup the container in the list of services in the docker-compose file
		dockerService := dockerCompose.Services[container]

		//update the shipment/container with the new image
		UpdateContainerImage(username, token, shipmentName, shipment, container, dockerService)

		//update container-level envvars
		for evName, evValue := range dockerService.Environment {

			if Verbose {
				log.Println("processing " + evName)
			}

			//create an envvar object
			envVarPayload := envVar(evName, evValue)

			//save the envvar
			SaveEnvVar(username, token, shipmentName, shipment, envVarPayload, container)

		} //envvars
	}

	//convert the specified barge into an env var
	if len(shipment.Barge) > 0 {

		//initialize the environment map if it doesn't exist
		if shipment.Environment == nil {
			shipment.Environment = make(map[string]string)
		}

		//set the BARGE env var
		shipment.Environment["BARGE"] = shipment.Barge
	}

	//update shipment/environment-level envvars
	for evName, evValue := range shipment.Environment {

		if Verbose {
			log.Println("processing " + evName)
		}

		//create an envvar object
		envVarPayload := envVar(evName, evValue)

		//save the envvar
		SaveEnvVar(username, token, shipmentName, shipment, envVarPayload, "")

	} //envvars

	//update shipment level configuration
	UpdateShipment(username, token, shipmentName, shipment)

	//trigger shipment
	_, messages := Trigger(shipmentName, shipment.Env)

	for _, msg := range messages {
		fmt.Println(msg)
	}

	//if replicas is changing from 0, then show wait messages
	if ec2Provider(currentShipment.Providers).Replicas == 0 {
		fmt.Println("Please allow up to 5 minutes for DNS changes to take effect.")
	}
}

func envVar(name string, value string) EnvVarPayload {
	return EnvVarPayload{
		Name:  name,
		Value: value,
		Type:  "basic",
	}
}
