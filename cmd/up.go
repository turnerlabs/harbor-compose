package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/docker/libcompose/project"
	"github.com/spf13/cobra"

	"strings"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start your application",
	Long: `Start your application
The up command applies changes from your docker/harbor compose files and brings your application up on Harbor.  The up command:

- Creates Harbor shipments if needed
- Updates container and shipment/environment level environment variables
- Updates and catalogs container images
- Updates container replicas
- Triggers your shipments
	`,
	Run: up,
}

func init() {
	RootCmd.AddCommand(upCmd)
}

var successMessage = "Please allow up to 5 minutes for Load Balancer and DNS changes to take effect."

const healthCheckEnvVarName = "HEALTHCHECK"

func up(cmd *cobra.Command, args []string) {

	//make sure user is authenticated
	username, token, err := Login()
	check(err)

	//read the compose files
	dockerCompose, harborCompose := unmarshalComposeFiles(DockerComposeFile, HarborComposeFile)

	//iterate shipments
	for shipmentName, shipment := range harborCompose.Shipments {
		if Verbose {
			log.Printf("processing shipment: %v/%v", shipmentName, shipment.Env)
		}

		//fetch the current state
		existingShipment := GetShipmentEnvironment(username, token, shipmentName, shipment.Env)

		//transform compose yaml into a desired NewShipmentEnvironment object
		desiredShipment := transformComposeToNewShipment(shipmentName, shipment, dockerCompose)

		//validate desired state
		err := validateUp(&desiredShipment, existingShipment)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(-1)
		}

		fmt.Printf("Starting %v %v ...\n", shipmentName, shipment.Env)

		//creating a shipment is a different workflow than updating
		if existingShipment == nil {
			if Verbose {
				log.Println("shipment environment not found")
			}
			createShipment(username, token, shipmentName, shipment, dockerCompose, desiredShipment)

		} else {
			//make changes to harbor based on compose files
			updateShipment(username, token, existingShipment, shipmentName, shipment, dockerCompose)
		}

		fmt.Println("done")

	} //shipments
}

//validates desire shipment against existing
func validateUp(desired *NewShipmentEnvironment, existing *ShipmentEnvironment) error {

	if Verbose {
		fmt.Println("existing:")
		b, e := json.Marshal(existing)
		check(e)
		fmt.Println(string(b))
		fmt.Println()
		fmt.Println("desired:")
		b, e = json.Marshal(desired)
		check(e)
		fmt.Println(string(b))
	}

	//env name
	if strings.Contains(desired.Environment.Name, "_") {
		if Verbose {
			fmt.Println(desired.Environment.Name)
		}
		return errors.New("environment can not contain underscores ('_')")
	}

	provider := ec2ProviderNewProvider(desired.Providers)

	//barge
	if provider.Barge == "" {
		return errors.New("barge is required for a shipment")
	}

	//replicas
	if Verbose {
		fmt.Println(provider.Replicas)
	}
	if !(provider.Replicas >= 0 && provider.Replicas <= 1000) {
		return errors.New("replicas must be between 1 and 1000")
	}

	//containers
	if len(desired.Containers) == 0 {
		return errors.New("at least 1 container is required")
	}

	for _, container := range desired.Containers {

		//ports
		if len(container.Ports) == 0 {
			return errors.New("At least one port is required.")
		}

		//validate health check
		foundHealthCheck := false
		for _, v := range container.Vars {
			if v.Name == healthCheckEnvVarName {
				foundHealthCheck = true
				break
			}
		}
		if !foundHealthCheck {
			return errors.New("A container-level 'HEALTHCHECK' environment variable is required")
		}
	}

	//update-specific validation
	if existing != nil {
		existingProvider := ec2Provider(existing.Providers)

		//don't allow barge changes
		if Verbose {
			fmt.Println("existing barge: " + existingProvider.Barge)
			fmt.Println("desired barge: " + provider.Barge)
		}
		if provider.Barge != existingProvider.Barge {
			return errors.New("Changing barges involves downtime. Please run the 'down' command first, then change barge and then run 'up' again.")
		}

		//don't allow container name changes
		for _, desiredContainer := range desired.Containers {
			//locate existing container with same name, error if not found
			found := false
			for _, existingContainer := range existing.Containers {
				if existingContainer.Name == desiredContainer.Name {

					//don't allow port changes
					existingPort := getPrimaryPort(existingContainer.Ports)
					desiredPort := getPrimaryPort(desiredContainer.Ports)
					if !(existingPort.Value == desiredPort.Value && existingPort.PublicPort == desiredPort.PublicPort) {
						return errors.New("Port changes involve downtime.  Please run the 'down --delete' command first.")
					}

					//don't allow health check changes
					if existingPort.Healthcheck != desiredPort.Healthcheck {
						return errors.New("Healthcheck changes involve downtime.  Please run the 'down --delete' command first.")
					}

					//return container match
					found = true
					break
				}
			}
			if !found {
				return errors.New("Container changes involve downtime.  Please run the 'down --delete' command first.")
			}
		}
	}

	return nil
}

//finds the primary port in a port slice
func getPrimaryPort(ports []PortPayload) PortPayload {
	for _, port := range ports {
		if port.Primary {
			return port
		}
	}
	return PortPayload{}
}

func transformComposeToNewShipment(shipmentName string, shipment ComposeShipment, dockerComposeProject project.APIProject) NewShipmentEnvironment {

	//create object used to create a new shipment environment from scratch
	newShipment := NewShipmentEnvironment{
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
	for containerIndex, container := range shipment.Containers {

		if Verbose {
			log.Printf("processing container: %v", container)
		}

		//lookup the container in the list of services in the docker-compose file
		serviceConfig := getDockerComposeService(dockerComposeProject, container)

		image := serviceConfig.Image
		if image == "" {
			log.Fatalln("'image' is required in docker compose file")
		}

		//parse image:tag and map to name/version
		parsedImage := strings.Split(image, ":")

		newContainer := NewContainer{
			Name:    container,
			Image:   image,
			Version: parsedImage[1],
			Vars:    make([]EnvVarPayload, 0),
			Ports:   make([]PortPayload, 0),
		}

		//container-level env vars (note that these are parsed by libcompose which supports:
		//environment, env_file, and variable substitution with .env)
		containerEnvVars := serviceConfig.Environment.ToMap()
		for name, value := range containerEnvVars {
			if name != "" {
				if Verbose {
					log.Println("processing " + name)
				}
				newContainer.Vars = append(newContainer.Vars, envVar(name, value))
			}
		}

		//map the docker compose service ports to harbor ports
		if len(serviceConfig.Ports) == 0 {
			log.Fatalln("At least one port mapping is required in docker compose file.")
		}

		//map first port in docker compose to default primary HTTP "PORT"
		parsedPort := strings.Split(serviceConfig.Ports[0], ":")

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
			Primary:     (containerIndex == 0),
			Protocol:    "http",
			External:    false,
			Healthcheck: containerEnvVars[healthCheckEnvVarName],
		}

		//add port to list
		newContainer.Ports = append(newContainer.Ports, primaryPort)

		//add container to list
		newShipment.Containers = append(newShipment.Containers, newContainer)
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

	return newShipment
}

func createShipment(username string, token string, shipmentName string, shipment ComposeShipment, dockerComposeProject project.APIProject, newShipment NewShipmentEnvironment) {

	if Verbose {
		log.Println("creating shipment environment")
	}

	//catalog containers
	for _, container := range shipment.Containers {
		serviceConfig := getDockerComposeService(dockerComposeProject, container)
		catalogContainer(container, serviceConfig.Image)
	}

	//push the new shipment/environment up to harbor
	SaveNewShipmentEnvironment(username, token, newShipment)

	//trigger shipment
	success, messages := Trigger(shipmentName, shipment.Env)

	for _, msg := range messages {
		fmt.Println(msg)
	}

	if success && shipment.Replicas > 0 {
		fmt.Println(successMessage)
	}
}

func updateShipment(username string, token string, currentShipment *ShipmentEnvironment, shipmentName string, shipment ComposeShipment, dockerComposeProject project.APIProject) {

	//map a ComposeShipment object (based on compose files) into
	//a series of API calls to update a shipment

	//iterate defined containers and apply container level updates
	for _, container := range shipment.Containers {
		if Verbose {
			log.Printf("processing container: %v", container)
		}

		//lookup the container in the list of services in the docker-compose file
		serviceConfig := getDockerComposeService(dockerComposeProject, container)

		// catalog containers
		catalogContainer(container, serviceConfig.Image)

		//update the shipment/container with the new image
		if !shipment.IgnoreImageVersion {
			UpdateContainerImage(username, token, shipmentName, shipment, container, serviceConfig.Image)
		}

		for evName, evValue := range serviceConfig.Environment.ToMap() {
			if evName != "" {
				if Verbose {
					log.Println("processing " + evName)
				}

				//create an envvar object
				envVarPayload := envVar(evName, evValue)

				//save the envvar
				SaveEnvVar(username, token, shipmentName, shipment, envVarPayload, container)
			}
		}

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
		fmt.Println(successMessage)
	}
}

func envVar(name string, value string) EnvVarPayload {
	return EnvVarPayload{
		Name:  name,
		Value: value,
		Type:  "basic",
	}
}

func catalogContainer(name string, image string) {

	if Verbose {
		log.Printf("cataloging container %v", name)
	}

	//parse image:tag and map to name/version
	parsedImage := strings.Split(image, ":")
	tag := parsedImage[1]

	//lookup container image in the catalog and catalog if missing
	if !IsContainerVersionCataloged(name, tag) {

		newContainer := CatalogitContainer{
			Name:    name,
			Image:   image,
			Version: tag,
		}

		message, err := Catalogit(newContainer)

		if Verbose {
			fmt.Println(message)
		}
		if err != nil {
			log.Fatal(err)
		}

	} else {
		if Verbose {
			log.Printf("container %v already cataloged", name)
		}
	}
}
