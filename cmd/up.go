package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/docker/libcompose/config"
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
		desiredShipment := transformComposeToShipmentEnvironment(shipmentName, shipment, dockerCompose)

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
func validateUp(desired *ShipmentEnvironment, existing *ShipmentEnvironment) error {

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
	if strings.Contains(desired.Name, "_") {
		if Verbose {
			fmt.Println(desired.Name)
		}
		return errors.New("environment can not contain underscores ('_')")
	}

	provider := ec2Provider(desired.Providers)

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
		for _, v := range container.EnvVars {
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

func transformComposeToShipmentEnvironment(shipmentName string, shipment ComposeShipment, dockerComposeProject project.APIProject) ShipmentEnvironment {

	//create object used to create a new shipment environment from scratch
	newShipment := ShipmentEnvironment{
		Name:    shipment.Env,
		EnvVars: make([]EnvVarPayload, 0),
		ParentShipment: ParentShipment{
			Name:    shipmentName,
			EnvVars: make([]EnvVarPayload, 0),
			Group:   shipment.Group,
		},
	}

	//add shipment-level env vars
	newShipment.ParentShipment.EnvVars = append(newShipment.ParentShipment.EnvVars, envVar("CUSTOMER", shipment.Group))
	newShipment.ParentShipment.EnvVars = append(newShipment.ParentShipment.EnvVars, envVar("PROPERTY", shipment.Property))
	newShipment.ParentShipment.EnvVars = append(newShipment.ParentShipment.EnvVars, envVar("PROJECT", shipment.Project))
	newShipment.ParentShipment.EnvVars = append(newShipment.ParentShipment.EnvVars, envVar("PRODUCT", shipment.Product))

	//default enableMonitoring to true if not specified in yaml
	enableMonitoring := true
	if shipment.EnableMonitoring != nil {
		enableMonitoring = *shipment.EnableMonitoring
	}
	newShipment.EnableMonitoring = enableMonitoring

	//add environment-level env vars
	for name, value := range shipment.Environment {
		newShipment.EnvVars = append(newShipment.EnvVars, envVar(name, value))
	}

	//containers
	//iterate defined containers and apply container level updates
	newShipment.Containers = make([]ContainerPayload, 0)
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

		newContainer := ContainerPayload{
			Name:    container,
			Image:   image,
			EnvVars: make([]EnvVarPayload, 0),
			Ports:   make([]PortPayload, 0),
		}

		//map docker-compose envvars to harbor env vars
		newContainer.EnvVars = transformDockerServiceEnvVarsToHarborEnvVars(serviceConfig)

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
			Healthcheck: getEnvVar(healthCheckEnvVarName, newContainer.EnvVars).Value,
		}

		//set the healthcheck values if they are defined in the yaml
		if shipment.HealthcheckTimeoutSeconds != nil {
			primaryPort.HealthcheckTimeout = shipment.HealthcheckTimeoutSeconds
		}
		if shipment.HealthcheckIntervalSeconds != nil {
			primaryPort.HealthcheckInterval = shipment.HealthcheckIntervalSeconds
		}

		//add port to list
		newContainer.Ports = append(newContainer.Ports, primaryPort)

		//add container to list
		newShipment.Containers = append(newShipment.Containers, newContainer)
	}

	//add default ec2 provider
	provider := ProviderPayload{
		Name:     "ec2",
		Barge:    shipment.Barge,
		Replicas: shipment.Replicas,
		EnvVars:  make([]EnvVarPayload, 0),
	}

	//add provider
	newShipment.Providers = append(newShipment.Providers, provider)

	return newShipment
}

func parseEnvVarNames(envFile string) []string {
	keys := []string{}

	//read the file
	contents, err := ioutil.ReadFile(envFile)
	check(err)

	//parse the file
	scanner := bufio.NewScanner(bytes.NewBuffer(contents))
	//iterate lines
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		//ignore comments and empty lines
		if len(line) > 0 && !strings.HasPrefix(line, "#") {
			key := strings.Split(line, "=")[0]
			keys = append(keys, key)
		}
	}
	return keys
}

func createShipment(username string, token string, shipmentName string, shipment ComposeShipment, dockerComposeProject project.APIProject, newShipment ShipmentEnvironment) {

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

		//should we process the image?
		if !shipment.IgnoreImageVersion {

			// catalog container image
			catalogContainer(container, serviceConfig.Image)

			//find the existing container
			currentContainer := findContainer(container, currentShipment.Containers)
			if currentContainer == nil {
				check(errors.New("Cannot find container. Adding new containers is not supported"))
			}

			//has the image changed?
			if serviceConfig.Image != currentContainer.Image {

				var payload = ContainerPayload{
					Name:  container,
					Image: serviceConfig.Image,
				}

				//update the shipment/container with the new image
				UpdateContainerImage(username, token, shipmentName, currentShipment.Name, payload)

			} else if Verbose {
				log.Println("image has not changed, skipping")
			}
		}

		//map docker-compose envvars to harbor env vars
		harborEnvVars := transformDockerServiceEnvVarsToHarborEnvVars(serviceConfig)
		for _, envvar := range harborEnvVars {
			if envvar.Name != "" {
				if Verbose {
					log.Printf("processing %s (%s)", envvar.Name, envvar.Type)
				}

				//TODO: check for delta against envvars in currentShipment
				//rather than doing additional GETs

				//save the envvar
				SaveEnvVar(username, token, shipmentName, shipment, envvar, container)
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

	//update settings related to ports
	updatePorts(currentShipment, shipment, username, token)

	//if user specified a value for enableMonitoring that's
	//different from current, then update
	if shipment.EnableMonitoring != nil && *shipment.EnableMonitoring != currentShipment.EnableMonitoring {
		if Verbose {
			fmt.Println("updating shipment/environment configuration (enableMonitoring)")
		}
		UpdateShipmentEnvironment(username, token, shipmentName, shipment)
	}

	//update provider configuration, if changed
	ec2 := ec2Provider(currentShipment.Providers)
	if shipment.Replicas != ec2.Replicas {

		providerPayload := ProviderPayload{
			Name:     ec2.Name,
			Replicas: shipment.Replicas,
		}

		UpdateProvider(username, token, shipmentName, currentShipment.Name, providerPayload)
	}

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

//update container ports
func updatePorts(existingShipment *ShipmentEnvironment, desiredShipment ComposeShipment, username string, token string) {

	//inspect container ports
	for _, container := range existingShipment.Containers {
		for _, port := range container.Ports {

			portPayload := UpdatePortRequest{
				Name: port.Name,
			}

			//only update the props that have been specified and changed

			if desiredShipment.HealthcheckTimeoutSeconds != nil && *desiredShipment.HealthcheckTimeoutSeconds != *port.HealthcheckTimeout {
				portPayload.HealthcheckTimeout = desiredShipment.HealthcheckTimeoutSeconds
			}

			if desiredShipment.HealthcheckIntervalSeconds != nil && *desiredShipment.HealthcheckIntervalSeconds != *port.HealthcheckInterval {
				portPayload.HealthcheckInterval = desiredShipment.HealthcheckIntervalSeconds
			}

			//do we need to send updates to the server?
			if portPayload.HealthcheckTimeout != nil || portPayload.HealthcheckInterval != nil {
				if Verbose {
					log.Printf("updating port: %s on container: %s\n", port.Name, container.Name)
				}
				updatePort(username, token, existingShipment.ParentShipment.Name, existingShipment.Name, container.Name, portPayload)
			}
		}
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

//transform a docker service's environment variables into harbor-specific env var objects
func transformDockerServiceEnvVarsToHarborEnvVars(dockerService *config.ServiceConfig) []EnvVarPayload {

	//docker-compose.yml
	//env_file:
	//- hidden.env
	//
	//gets mapped to type=hidden
	//everything else type=basic

	harborEnvVars := []EnvVarPayload{}

	//container-level env vars (note that these are parsed by libcompose which supports:
	//environment, env_file, and variable substitution with .env)
	containerEnvVars := dockerService.Environment.ToMap()

	//has the user specified hidden env vars in a hidden.env?
	hiddenEnvVars := false
	hiddenEnvVarFile := ""
	for _, envFileName := range dockerService.EnvFile {
		if strings.HasSuffix(envFileName, hiddenEnvFileName) {
			hiddenEnvVars = true
			hiddenEnvVarFile = envFileName
			break
		}
	}

	//iterate/process hidden envvars and remove them from the list
	if hiddenEnvVars {
		if Verbose {
			log.Println("found hidden env vars")
		}
		for _, name := range parseEnvVarNames(hiddenEnvVarFile) {
			if Verbose {
				log.Println("processing " + name)
			}
			harborEnvVars = append(harborEnvVars, envVarHidden(name, containerEnvVars[name]))
			delete(containerEnvVars, name)
		}
	}

	//iterate/process envvars (hidden have already filtered out)
	for name, value := range containerEnvVars {
		if name != "" {
			if Verbose {
				log.Println("processing " + name)
			}
			harborEnvVars = append(harborEnvVars, envVar(name, value))
		}
	}

	return harborEnvVars
}
