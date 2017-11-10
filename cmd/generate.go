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
	Short: "Generate compose files, build artifacts, and terraform source from an existing shipment",
	Long: `Generate compose files, build artifacts, and terraform source from an existing shipment

The generate command outputs compose files, build artifacts, and terraform source that allow you to build and run your app locally in Docker, manage your Harbor infrastructure using Terraform, do CI/CD, and deploy images and environment variables.

Example:
harbor-compose generate my-shipment dev

The generate command's --build-provider flag allows you to generate build provider-specific files that allow you to build Docker images and do CI/CD with Harbor.

Examples:
harbor-compose generate my-shipment dev
harbor-compose generate my-shipment dev --build-provider local
harbor-compose generate my-shipment dev -b circleciv1
harbor-compose generate my-shipment dev -b circleciv2
harbor-compose generate my-shipment dev -b codeship
harbor-compose generate my-shipment dev --terraform
`,
	Run: generate,
}

var buildProvider string
var generateTerraform bool

const (
	providerEc2       = "ec2"
	hiddenEnvFileName = "hidden.env"
)

func init() {
	generateCmd.PersistentFlags().StringVarP(&buildProvider, "build-provider", "b", "", "generate build provider-specific files that allow you to build Docker images do CI/CD with Harbor")

	generateCmd.PersistentFlags().BoolVarP(&generateTerraform, "terraform", "t", false, "generate a Terraform source file (main.tf) that allows you to start managing your shipment using Terraform")

	RootCmd.AddCommand(generateCmd)
}

func generate(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		log.Fatal("at least 2 arguments are required. ex: harbor-compose generate my-shipment dev")
	}

	username, token, err := Login()
	check(err)

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

	//convert a Shipment object into a HarborCompose object
	harborCompose, hiddenEnvVars := transformShipmentToHarborCompose(shipmentObject)

	//convert a Shipment object into a DockerCompose object, with hidden envvars
	dockerCompose := transformShipmentToDockerCompose(shipmentObject, hiddenEnvVars)

	//if build provider is specified, allow it modify the compose objects and do its thing
	if len(buildProvider) > 0 {
		provider, err := getBuildProvider(buildProvider)
		check(err)

		artifacts, err := provider.ProvideArtifacts(&dockerCompose, &harborCompose, shipmentObject.BuildToken)
		check(err)

		//write artifacts to file system
		if artifacts != nil {
			for _, artifact := range artifacts {
				//create directories if needed
				dirs := filepath.Dir(artifact.FilePath)
				err = os.MkdirAll(dirs, os.ModePerm)
				check(err)

				if _, err := os.Stat(artifact.FilePath); err == nil {
					//exists
					fmt.Print(artifact.FilePath + " already exists. Overwrite? ")
					if askForConfirmation() {
						err = ioutil.WriteFile(artifact.FilePath, []byte(artifact.FileContents), artifact.FileMode)
						check(err)
					}
				} else {
					//doesn't exist
					err = ioutil.WriteFile(artifact.FilePath, []byte(artifact.FileContents), artifact.FileMode)
					check(err)
				}
			}
		}
	}

	//prompt if the file already exists
	yes := true
	if _, err := os.Stat(DockerComposeFile); err == nil {
		fmt.Print("docker-compose.yml already exists. Overwrite? ")
		yes = askForConfirmation()
	}
	if yes {
		SerializeDockerCompose(dockerCompose, DockerComposeFile)
		fmt.Println("wrote " + DockerComposeFile)
	}

	//prompt if the file already exist
	if _, err := os.Stat(HarborComposeFile); err == nil {
		fmt.Print("harbor-compose.yml already exists. Overwrite? ")
		yes = askForConfirmation()
	}
	if yes {
		SerializeHarborCompose(harborCompose, HarborComposeFile)
		fmt.Println("wrote " + HarborComposeFile)
	}

	if len(hiddenEnvVars) > 0 {

		//prompt to override hidden env file
		if _, err := os.Stat(hiddenEnvFileName); err == nil {
			fmt.Print(hiddenEnvFileName + " already exists. Overwrite? ")
			yes = askForConfirmation()
		}
		if yes {
			writeHiddenEnvFile(hiddenEnvVars, hiddenEnvFileName)
			fmt.Println("wrote " + hiddenEnvFileName)
		}

		//add hidden env_file to .gitignore and .dockerignore (to avoid checking secrets)
		sensitiveFiles := []string{hiddenEnvFileName, ".terraform"}
		appendToFile(".gitignore", sensitiveFiles)
		appendToFile(".dockerignore", sensitiveFiles)
	}

	//if the --terraform flag is specified, output a main.tf file
	if generateTerraform {
		generateAndWriteTerraformSource(shipmentObject, &harborCompose, true)
	}

	fmt.Println("done")
}

func writeHiddenEnvFile(envvars map[string]string, file string) {
	contents := ""
	for name, value := range envvars {
		contents += fmt.Sprintf("%s=%s\n", name, value)
	}
	err := ioutil.WriteFile(file, []byte(contents), 0644)
	check(err)
}

func transformShipmentToHarborCompose(shipmentObject *ShipmentEnvironment) (HarborCompose, map[string]string) {

	//convert a Shipment object into a HarborCompose object with a single shipment
	harborCompose := HarborCompose{
		Shipments: make(map[string]ComposeShipment),
	}

	//lookup primary port
	primaryPort := getPrimaryPort(shipmentObject.Containers[0].Ports)

	composeShipment := ComposeShipment{
		Env:                        shipmentObject.Name,
		Group:                      shipmentObject.ParentShipment.Group,
		Environment:                make(map[string]string),
		EnableMonitoring:           &shipmentObject.EnableMonitoring,
		HealthcheckTimeoutSeconds:  primaryPort.HealthcheckTimeout,
		HealthcheckIntervalSeconds: primaryPort.HealthcheckInterval,
	}

	//track special envvars
	special := map[string]string{}

	//track hidden envvars
	hiddenEnvVars := map[string]string{}

	//shipment
	copyEnvVars(shipmentObject.ParentShipment.EnvVars, nil, special, hiddenEnvVars)

	//environment
	copyEnvVars(shipmentObject.EnvVars, nil, special, hiddenEnvVars)

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
	for _, container := range shipmentObject.Containers {
		composeShipment.Containers = append(composeShipment.Containers, container.Name)
	}

	//add single shipment to list
	harborCompose.Shipments[shipmentObject.ParentShipment.Name] = composeShipment

	return harborCompose, hiddenEnvVars
}

//transforms a ShipmentEnvironment object to its DockerCompose representation
//(along with hidden env vars)
func transformShipmentToDockerCompose(shipmentObject *ShipmentEnvironment, hiddenEnvVars map[string]string) DockerCompose {

	dockerCompose := DockerCompose{
		Version:  "2",
		Services: make(map[string]*DockerComposeService),
	}

	//convert containers to docker services
	for _, container := range shipmentObject.Containers {

		//create a docker service based on this container
		service := DockerComposeService{
			Image:       container.Image,
			Ports:       []string{},
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
		copyEnvVars(shipmentObject.ParentShipment.EnvVars, service.Environment, nil, hiddenEnvVars)

		//environment
		copyEnvVars(shipmentObject.EnvVars, service.Environment, nil, hiddenEnvVars)

		//container
		copyEnvVars(container.EnvVars, service.Environment, nil, hiddenEnvVars)

		//write hidden env vars to file specified in env_file
		if len(hiddenEnvVars) > 0 {
			service.EnvFile = []string{hiddenEnvFileName}
		}

		//add service to list
		dockerCompose.Services[container.Name] = &service
	}

	return dockerCompose
}
