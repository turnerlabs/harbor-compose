package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate compose files and build artifacts from an existing shipment",
	Long: `Generate compose files and build artifacts from an existing shipment

The generate command outputs compose files and build artifacts that allow you to build and run your app locally in Docker, do CI/CD in Harbor, and make changes in Harbor using the up command.

Example:
harbor-compose generate my-shipment dev

The generate command's --build-provider flag allows you to generate build provider-specific files that allow you to build Docker images and do CI/CD with Harbor.

Examples:
harbor-compose generate my-shipment dev --build-provider local
harbor-compose generate my-shipment dev -b circleciv1
harbor-compose generate my-shipment dev -b circleciv2
`,
	Run: generate,
}

var buildProvider string

func init() {
	generateCmd.PersistentFlags().StringVarP(&buildProvider, "build-provider", "b", "", "generate build provider-specific files that allow you to build Docker images do CI/CD with Harbor")
	RootCmd.AddCommand(generateCmd)
}

func transformShipmentToDockerCompose(shipmentObject *ShipmentEnvironment) DockerCompose {

	dockerCompose := DockerCompose{
		Version:  "2",
		Services: make(map[string]*DockerComposeService),
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

		//container
		copyEnvVars(container.EnvVars, service.Environment, nil)

		//add service to list
		dockerCompose.Services[container.Name] = &service
	}

	return dockerCompose
}

func generate(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		log.Fatal("at least 2 arguments are required. ex: harbor-compose generate my-shipment dev")
	}

	username, token, err := Login()
	if err != nil {
		log.Fatal(err)
	}

	shipment := args[0]
	env := args[1]

	if Verbose {
		log.Printf("fetching shipment...")
	}
	shipmentObject := GetShipmentEnvironment(username, token, shipment, env)
	if shipmentObject == nil {
		fmt.Println("shipment not found")
		return
	}

	//convert a Shipment object into a DockerCompose object
	dockerCompose := transformShipmentToDockerCompose(shipmentObject)

	//convert a Shipment object into a HarborCompose object
	harborCompose := transformShipmentToHarborCompose(shipmentObject, &dockerCompose)

	//if build provider is specified, allow it modify the compose objects and do its thing
	if len(buildProvider) > 0 {
		provider, err := getBuildProvider(buildProvider)
		if err != nil {
			log.Fatal(err)
		}
		artifacts, err := provider.ProvideArtifacts(&dockerCompose, &harborCompose, shipmentObject.BuildToken)
		if err != nil {
			log.Fatal(err)
		}

		//write artifacts to file system
		if artifacts != nil {
			for _, artifact := range artifacts {
				//create directories if needed
				dirs := filepath.Dir(artifact.FilePath)
				err = os.MkdirAll(dirs, os.ModePerm)
				if err != nil {
					log.Fatal(err)
				}
				if _, err := os.Stat(artifact.FilePath); err == nil {
					//exists
					fmt.Print(artifact.FilePath + " already exists. Overwrite? ")
					if askForConfirmation() {
						err = ioutil.WriteFile(artifact.FilePath, []byte(artifact.FileContents), 0644)
					}
				} else {
					//doesn't exist
					err = ioutil.WriteFile(artifact.FilePath, []byte(artifact.FileContents), 0644)
				}
			}
		}
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

func transformShipmentToHarborCompose(shipmentObject *ShipmentEnvironment, dockerCompose *DockerCompose) HarborCompose {

	//convert a Shipment object into a HarborCompose object with a single shipment
	harborCompose := HarborCompose{
		Shipments: make(map[string]ComposeShipment),
	}

	composeShipment := ComposeShipment{
		Env:         shipmentObject.Name,
		Group:       shipmentObject.ParentShipment.Group,
		Environment: make(map[string]string),
	}

	//track special envvars
	special := map[string]string{}

	//shipment
	copyEnvVars(shipmentObject.ParentShipment.EnvVars, nil, special)

	//environment
	copyEnvVars(shipmentObject.EnvVars, nil, special)

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

	return harborCompose
}
