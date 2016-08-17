package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate docker-compose.yml and harbor-compose.yml files from an existing shipment",
	Long: `
$ harbor-compose generate my-shipment dev
or
$ harbor-compose generate my-shipment qa --user foo
`,
	Run: generate,
}

func init() {
	RootCmd.AddCommand(generateCmd)
}

func generate(cmd *cobra.Command, args []string) {

	if len(args) < 2 {
		log.Fatal("2 arguments are required. ex: harbor-compose generate my-shipment dev")
	}

	_, _, err := Login()
	if err != nil {
		log.Fatalf(err.Error())
	}

	shipment := args[0]
	env := args[1]

	if Verbose {
		log.Printf("fetching shipment...")
	}
	shipmentObject := GetShipmentEnvironment(shipment, env)

	//convert a Shipment object into a DockerCompose object
	dockerCompose := DockerCompose{
		Version:  "2",
		Services: make(map[string]DockerComposeService),
	}

	//convert containers to docker services
	for _, container := range shipmentObject.Containers {

		//create a docker service based on this container
		service := DockerComposeService{
			Image:       container.Image,
			Ports:       make([]string, len(shipmentObject.Ports)),
			Environment: make(map[string]string),
		}

		//populate ports
		for _, port := range container.Ports {

			//format = external:internal
			if port.PublicPort == 0 {
				port.PublicPort = port.Value
			}
			dockerPort := fmt.Sprintf("%v:%v", port.PublicPort, port.Value)
			service.Ports = append(service.Ports, dockerPort)

			//set container env vars for healthcheck, and port
			//so that apps can simulate running in harbor
			service.Environment["PORT"] = strconv.Itoa(port.Value)
			service.Environment["HEALTHCHECK"] = port.Healthcheck
		}

		//copy shipment, environment, provider level env vars down to the
		//container level so that they can be used in docker-compose

		//shipment
		copyEnvVars(shipmentObject.ParentShipment.EnvVars, service.Environment, nil)

		//environment
		copyEnvVars(shipmentObject.EnvVars, service.Environment, nil)

		//todo: provider

		//container
		copyEnvVars(container.EnvVars, service.Environment, nil)

		//add service to list
		dockerCompose.Services[container.Name] = service
	}

	//prompt if the file already exist
	if _, err := os.Stat(DockerComposeFile); err == nil {
		//exists
		fmt.Print("docker-compose.yml already exists. Overwrite? ")
		if askForConfirmation() {
			SerializeDockerCompose(dockerCompose, DockerComposeFile)
		}
	} else {
		//doesn't exist
		SerializeDockerCompose(dockerCompose, DockerComposeFile)
	}

	//now generate harbor-compose.yml

	//convert a Shipment object into a HarborCompose object with a single shipment
	harborCompose := HarborCompose{
		Shipments: make(map[string]ComposeShipment),
	}

	composeShipment := ComposeShipment{
		Env:         shipmentObject.Name,
		Group:       shipmentObject.ParentShipment.Group,
		Environment: make(map[string]string),
	}

	//populate env vars

	//track special envvars
	special := map[string]string{}

	//shipment
	copyEnvVars(shipmentObject.ParentShipment.EnvVars, composeShipment.Environment, special)

	//environment
	copyEnvVars(shipmentObject.EnvVars, composeShipment.Environment, special)

	//todo: provider envvars

	//look for the ec2 provider (for now)
	provider := ec2Provider(shipmentObject.Providers)

	//now populate other harbor-compose metadata
	composeShipment.Product = special["PRODUCT"]
	composeShipment.Project = special["PROJECT"]
	composeShipment.Property = special["PROPERTY"]

	//use the barge setting on the provider, otherwise use the envvar
	composeShipment.Barge = provider.Barge
	if composeShipment.Barge == "" {
		composeShipment.Barge = special["BARGE"]
	}

	//set replicas from the provider
	composeShipment.Replicas = provider.Replicas

	//add containers
	for container := range dockerCompose.Services {
		composeShipment.Containers = append(composeShipment.Containers, container)
	}

	//add single shipment to list
	harborCompose.Shipments[shipmentObject.ParentShipment.Name] = composeShipment

	//prompt if the file already exist
	if _, err := os.Stat(HarborComposeFile); err == nil {
		//exists
		fmt.Print("harbor-compose.yml already exists. Overwrite? ")
		if askForConfirmation() {
			SerializeHarborCompose(harborCompose, HarborComposeFile)
			fmt.Println("done")
		}
	} else {
		//doesn't exist
		SerializeHarborCompose(harborCompose, HarborComposeFile)
		fmt.Println("done")
	}

}

//find the ec2 provider
func ec2Provider(providers []ProviderPayload) *ProviderPayload {
	for _, provider := range providers {
		if provider.Name == "ec2" {
			return &provider
		}
	}
	log.Fatal("ec2 provider is missing")
	return nil
}
