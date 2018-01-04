package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/docker/libcompose/config"
	"github.com/spf13/cobra"
)

const (

	//special
	envVarNameCustomer = "CUSTOMER"
	envVarNameProduct  = "PRODUCT"
	envVarNameProject  = "PROJECT"
	envVarNameProperty = "PROPERTY"
	envVarNameBarge    = "BARGE"
	envVarNameRestart  = "HC_RESTART"

	//log shipping
	envVarNameShipLogs     = "SHIP_LOGS"
	envVarNameLogsEndpoint = "LOGS_ENDPOINT"
	envVarNameAccessKey    = "LOGS_ACCESS_KEY"
	envVarNameSecretKey    = "LOGS_SECRET_KEY"
	envVarNameDomainName   = "LOGS_DOMAIN_NAME"
	envVarNameRegion       = "LOGS_REGION"
	envVarNameQueueName    = "LOGS_QUEUE_NAME"
)

func specialEnvVars() map[string]string {
	return map[string]string{
		envVarNameCustomer: envVarNameCustomer,
		envVarNameProduct:  envVarNameProduct,
		envVarNameProject:  envVarNameProject,
		envVarNameProperty: envVarNameProperty,
		envVarNameBarge:    envVarNameBarge,
		envVarNameRestart:  envVarNameRestart,
	}
}

func logShippingEnvVars() map[string]string {
	return map[string]string{
		envVarNameShipLogs:     envVarNameShipLogs,
		envVarNameLogsEndpoint: envVarNameLogsEndpoint,
		envVarNameAccessKey:    envVarNameAccessKey,
		envVarNameSecretKey:    envVarNameSecretKey,
		envVarNameDomainName:   envVarNameDomainName,
		envVarNameRegion:       envVarNameRegion,
		envVarNameQueueName:    envVarNameQueueName,
	}
}

var envShipment string
var envEnvironment string
var envHiddenFile string
var envEnvFile string

func init() {
	RootCmd.AddCommand(envCmd)

	//list
	envCmd.AddCommand(listEnvCmd)
	listEnvCmd.PersistentFlags().StringVarP(&envShipment, "shipment", "s", "", "shipment name")
	listEnvCmd.PersistentFlags().StringVarP(&envEnvironment, "environment", "e", "", "environment name")

	//push
	envCmd.AddCommand(pushEnvCmd)
	pushEnvCmd.PersistentFlags().StringVarP(&envShipment, "shipment", "s", "", "shipment name")
	pushEnvCmd.PersistentFlags().StringVarP(&envEnvironment, "environment", "e", "", "environment name")
	pushEnvCmd.PersistentFlags().StringVarP(&envHiddenFile, "hidden", "", hiddenEnvFileName, "The location of the docker compose environment file that contains hidden environment variables")

	//pull
	envCmd.AddCommand(pullEnvCmd)
	pullEnvCmd.PersistentFlags().StringVarP(&envShipment, "shipment", "s", "", "shipment name")
	pullEnvCmd.PersistentFlags().StringVarP(&envEnvironment, "environment", "e", "", "environment name")
	pullEnvCmd.PersistentFlags().StringVarP(&envHiddenFile, "hidden", "", hiddenEnvFileName, "The location of the docker compose environment file that contains hidden environment variables")
	pullEnvCmd.PersistentFlags().StringVarP(&envEnvFile, "env-file", "", "", "Specify a docker compose env_file to write to rather than writing directly to the docker-compose.yml environment section")
}

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "manage environment variables",
	Long:  "manage environment variables",
	Example: `harbor-compose env list
harbor-compose env push 
harbor-compose env pull`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	PreRun: preRunHook,
}

var listEnvCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list harbor environment variables",
	Long: `list harbor environment variables

List environment and container-level environment variables for all shipment/environments in harbor-compose.yml

Use the --shipment and --environment flags to specify a shipment/environment other than what's in the harbor-compose.yml
`,
	Example: `harbor-compose env ls
harbor-compose env ls -s my-app -e dev
`,
	Run:    listEnvVars,
	PreRun: preRunHook,
}

// pushEnvCmd represents the envvars push command
var pushEnvCmd = &cobra.Command{
	Use:   "push",
	Short: "push docker compose environment variables to harbor",
	Long: `push docker compose environment variables to harbor

The push command takes all of the environment variables accessible by docker-compose and uploads them to Harbor.  Note that this command does not trigger a deployment.

The push command works with a harbor-compose.yml file to push environment variables for one or many shipment/environment/containers, as well as for a single shipment environment using the --shipment and --environment flags.
`,
	Example: `harbor-compose env push
harbor-compose env push -s my-shipment -e dev

You can specify which env file contains your hidden environment variables using the --hidden flag (defaults to hidden.env)
harbor-compose env push --hidden secrets.env
`,
	Run:    pushEnvVars,
	PreRun: preRunHook,
}

// pullEnvCmd represents the envvars pull command
var pullEnvCmd = &cobra.Command{
	Use:   "pull",
	Short: "pull harbor environment variables into docker compose",
	Long: `pull harbor environment variables into docker compose

The pull command fetches environment variables from harbor for each of the shipment/environment/containers specified in harbor-compose.yml and write them to docker compose files.

By default, non-hidden env vars are written directly to the environment section of docker-compose.yml and hidden env vars are written to hidden.env. The optional --env-file flag indicates that non-hidden env vars should be written to this file and reflected in the env_file section of docker-compose.yml.  Similarly, the --hidden flag indicates that hidden env vars should be written to this file and reflected in the env_file section of docker-compose.yml.

The pull command also takes optional --shipment and --environment flags.
`,
	Example: `harbor-compose env pull 
harbor-compose env pull -s my-app -e dev

You can use the optional --env-file and --hidden flags to specify where the environment variables get written to.
harbor-compose env pull --env-file public.env --hidden private.env

Specify the shipment/environment using the --shipment and --environment flags instead of a harbor-compose.yml file
harbor-compose env pull --shipment my-app --environment dev
harbor-compose env pull -s my-app -e dev
`,
	Run:    pullEnvVars,
	PreRun: preRunHook,
}

func listEnvVars(cmd *cobra.Command, args []string) {

	//make sure user is authenticated
	username, token, err := Login()
	check(err)

	//determine which shipment/environments user wants to process
	inputShipmentEnvironments, _ := getShipmentEnvironmentsFromInput(envShipment, envEnvironment)

	//iterate shipment/environments
	for _, shipmentEnv := range inputShipmentEnvironments {
		shipment := shipmentEnv.Item1
		env := shipmentEnv.Item2

		fmt.Printf("SHIPMENT: %v\n", shipment)
		fmt.Printf("ENVIRONMENT: %v\n", env)

		//lookup the shipment environment
		shipmentEnvironment := GetShipmentEnvironment(username, token, shipment, env)
		if shipmentEnvironment == nil {
			fmt.Println(messageShipmentEnvironmentNotFound)
			return
		}

		//filter out special env vars
		envvarsToPrint := []EnvVarPayload{}
		for _, envvar := range shipmentEnvironment.EnvVars {
			if specialEnvVars()[envvar.Name] == "" {
				envvarsToPrint = append(envvarsToPrint, envvar)
			}
		}
		if len(envvarsToPrint) > 0 {
			printEnvVars(envvarsToPrint)
		} else {
			fmt.Println()
		}

		//iterate containers
		for _, container := range shipmentEnvironment.Containers {
			fmt.Println("CONTAINER: " + container.Name)
			printEnvVars(container.EnvVars)
		}
	}
}

func printEnvVars(envvars []EnvVarPayload) {
	//print table header
	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.DiscardEmptyColumns)
	fmt.Println("")
	fmt.Fprintln(w, "NAME\tVALUE\tTYPE")

	//create a formatted template
	tmpl, err := template.New("envvars").Parse("{{.Name}}\t{{.Value}}\t{{.Type}}")
	check(err)

	//print non-special env vars
	for _, envvar := range envvars {
		if specialEnvVars()[envvar.Name] == "" {
			//execute the template with the data
			err = tmpl.Execute(w, envvar)
			check(err)
			fmt.Fprintln(w)
		}
	}

	//flush the writer
	w.Flush()
	fmt.Println("")
}

func pushEnvVars(cmd *cobra.Command, args []string) {

	//make sure user is authenticated
	username, token, err := Login()
	check(err)

	//determine which shipment/environments user wants to process
	inputShipmentEnvironments, harborComposeConfig := getShipmentEnvironmentsFromInput(envShipment, envEnvironment)

	//load docker compose file
	dc := DeserializeDockerCompose(DockerComposeFile)

	//iterate shipment/environments
	for _, t := range inputShipmentEnvironments {
		shipment := t.Item1
		env := t.Item2

		//lookup the shipment environment
		shipmentEnvironment := GetShipmentEnvironment(username, token, shipment, env)
		if shipmentEnvironment == nil {
			fmt.Println(messageShipmentEnvironmentNotFound)
			return
		}

		//iterate containers
		for _, container := range shipmentEnvironment.Containers {
			if Verbose {
				log.Printf("processing container: %v", container.Name)
			}

			//lookup the container in the list of services in the docker-compose file
			serviceConfig := getDockerComposeService(dc, container.Name)

			//translate docker envvars to harbor
			harborEnvVars := transformDockerServiceEnvVarsToHarborEnvVarsHidden(serviceConfig, envHiddenFile)
			for _, envvar := range harborEnvVars {
				if envvar.Name != "" {
					if Verbose {
						log.Printf("processing %s (%s)", envvar.Name, envvar.Type)
					}

					//save the envvar
					SaveEnvVar(username, token, shipment, env, envvar, container.Name)
				}
			}
		}

		//only process environment-level env vars if using a harbor-compose.yml file
		if harborComposeConfig != nil {
			if Verbose {
				log.Println("looking for harbor-compose.yml envvars")
			}

			//lookup current shipment in harbor-compose.yml
			for evName, evValue := range harborComposeConfig.Shipments[shipment].Environment {
				if Verbose {
					log.Println("processing " + evName)
				}

				//translate to a a harbor basic env var
				envVarPayload := envVar(evName, evValue)

				//save the envvar
				SaveEnvVar(username, token, shipment, env, envVarPayload, "")

			} //envvars
		}
	}

	fmt.Println("done")
	fmt.Println("run 'up' or 'deploy' for the environment variable changes to take effect")
}

func pullEnvVars(cmd *cobra.Command, args []string) {

	//make sure user is authenticated
	username, token, err := Login()
	check(err)

	//determine which shipment/environments user wants to process
	inputShipmentEnvironments, localHarborCompose := getShipmentEnvironmentsFromInput(envShipment, envEnvironment)

	//build up a DockerCompose object that we'll use for outputting docker-compose.yml
	//this object may get services/containers from multiple shipment/environments
	//(depending on harbor-compose.yml)
	pulledDockerCompose := DockerCompose{
		Version:  "2",
		Services: map[string]*DockerComposeService{},
	}

	//track env_files that need to be written
	hiddenEnvVars := make(map[string]string)
	nonHiddenEnvVars := make(map[string]string)

	//iterate shipment/environments
	for _, t := range inputShipmentEnvironments {
		shipment := t.Item1
		env := t.Item2

		//fetch the shipment environment from the backend
		shipmentEnvironment := GetShipmentEnvironment(username, token, shipment, env)
		if shipmentEnvironment == nil {
			fmt.Println(messageShipmentEnvironmentNotFound)
			return
		}

		//convert shipit shipment/env object into a new HarborCompose object
		//representing the remote state
		remoteHarborCompose := transformShipmentToHarborCompose(shipmentEnvironment)

		//update local harbor-compose object (used for writing) environment with remote envvars
		remoteEnvironmentEnvVars := remoteHarborCompose.Shipments[shipment].Environment
		if localHarborCompose != nil && len(remoteEnvironmentEnvVars) > 0 {
			if Verbose {
				fmt.Printf("updating %s with remote environment-level envvars \n", HarborComposeFile)
			}
			currentShipment := localHarborCompose.Shipments[shipment]
			currentShipment.Environment = remoteEnvironmentEnvVars
			localHarborCompose.Shipments[shipment] = currentShipment
		}

		//now translate remote harbor shipment/environnment into a DockerCompose object
		//add the services/containers from this shipment/env to the pulled compose object
		//non-hidden envvars will get written to docker-compose.yml or --env-file
		//hidden envvars will get written to --hidden
		remoteDockerCompose, remoteHiddenEnvVars := transformShipmentToDockerComposeWithEnvFile(shipmentEnvironment, envHiddenFile)

		//track hidden envvars to be written later
		for k, v := range remoteHiddenEnvVars {
			hiddenEnvVars[k] = v
		}

		//iterate containers
		for name, service := range remoteDockerCompose.Services {

			//if --env-file is specified, write non-hidden envvars there and update env_file
			//otherwise, they will get written directly to docker-compose.yml
			if envEnvFile != "" {
				for k, v := range service.Environment {
					nonHiddenEnvVars[k] = v
				}

				//clear out environment section and add pointer to env_file
				service.Environment = make(map[string]string)
				service.EnvFile = append(service.EnvFile, envEnvFile)
			}

			//add this service to master compose file
			pulledDockerCompose.Services[name] = service
		}
	} //shipment/env to process

	//output harbor-compose.yml, if not using -s -e
	if localHarborCompose != nil {
		content := marshalHarborCompose(*localHarborCompose)
		outputFile(string(content), HarborComposeFile)
	}

	//output docker-compose.yml
	content := marshalDockerCompose(pulledDockerCompose)
	outputFile(string(content), DockerComposeFile)

	//write hidden env vars to --hidden
	outputEnvFile(hiddenEnvVars, envHiddenFile)

	//write non-hidden envvars to --env-file
	outputEnvFile(nonHiddenEnvVars, envEnvFile)

	fmt.Println("done")
}

func serializeToEnvFile(envvars map[string]string) string {
	//extract key slice
	keys := []string{}
	for key := range envvars {
		keys = append(keys, key)
	}

	//sort
	sort.Strings(keys)

	//write to env_file format (key=value)
	result := ""
	for _, key := range keys {
		result += fmt.Sprintf("%s=%s\n", key, envvars[key])
	}
	return result
}

func outputFile(content string, file string) {

	//does the file already exist?
	writeFile := true
	if _, err := os.Stat(file); err == nil {

		//read existing file
		b, err := ioutil.ReadFile(file)
		check(err)
		existingContent := string(b)

		//has the file changed?
		if content != existingContent {
			fmt.Print(file + " already exists. Overwrite? ")
			writeFile = askForConfirmation()
		} else {
			writeFile = false
			if Verbose {
				fmt.Printf("%s hasn't changed \n", file)
			}
		}
	}

	if writeFile {
		err := ioutil.WriteFile(file, []byte(content), 0644)
		check(err)
		fmt.Println("wrote " + file)
	}
}

func outputEnvFile(envvars map[string]string, file string) {
	if len(envvars) > 0 {

		//serialize envvars map to env_file format (sorted by key)
		newEnvFile := serializeToEnvFile(envvars)

		//does the file already exist?
		writeFile := true
		if _, err := os.Stat(file); err == nil {

			//read existing file
			b, err := ioutil.ReadFile(file)
			check(err)
			oldEnvFile := string(b)

			//do diff and see if the contents have changed
			if Verbose {
				fmt.Println("old")
				fmt.Println(oldEnvFile)
				fmt.Println("new")
				fmt.Println(newEnvFile)
			}
			if newEnvFile != oldEnvFile {
				//prompt to override env file
				fmt.Print(file + " already exists. Overwrite? ")
				writeFile = askForConfirmation()
			} else {
				writeFile = false
				if Verbose {
					fmt.Printf("%s hasn't changed \n", file)
				}
			}
		}

		if writeFile {
			err := ioutil.WriteFile(file, []byte(newEnvFile), 0644)
			check(err)
			fmt.Println("wrote " + file)
		}
	}
}

//processes envvars by copying them to a destination and filtering out special and hidden envvars
func copyEnvVars(source []EnvVarPayload, destination map[string]string, special map[string]string, hidden map[string]string, logShipping map[string]string) {

	//iterate all envvars
	for _, envvar := range source {

		//is this a special envvar?
		if specialEnvVars()[strings.ToUpper(envvar.Name)] == "" { //no

			//escape `$` characters with `$$`
			envvar.Value = strings.Replace(envvar.Value, "$", "$$", -1)

			if logShippingEnvVars()[strings.ToUpper(envvar.Name)] != "" {
				if logShipping != nil {
					logShipping[envvar.Name] = envvar.Value
				}
			} else if envvar.Type == "hidden" && hidden != nil {
				//copy to hidden
				hidden[envvar.Name] = envvar.Value

			} else if destination != nil {
				//copy to destination
				destination[envvar.Name] = envvar.Value
			}
		} else if special != nil {
			//copy to special
			special[envvar.Name] = envvar.Value
		}
	}
}

func getEnvVar(name string, vars []EnvVarPayload) *EnvVarPayload {
	for _, envvar := range vars {
		if envvar.Name == name {
			return &envvar
		}
	}
	return &EnvVarPayload{}
}

func envVar(name string, value string) EnvVarPayload {
	return EnvVarPayload{
		Name:  name,
		Value: value,
		Type:  "basic",
	}
}

func envVarHidden(name string, value string) EnvVarPayload {
	return EnvVarPayload{
		Name:  name,
		Value: value,
		Type:  "hidden",
	}
}

//transform a docker service's environment variables into harbor-specific env var objects
func transformDockerServiceEnvVarsToHarborEnvVars(dockerService *config.ServiceConfig) []EnvVarPayload {
	return transformDockerServiceEnvVarsToHarborEnvVarsHidden(dockerService, hiddenEnvFileName)
}

//transform a docker service's environment variables into harbor-specific env var objects
func transformDockerServiceEnvVarsToHarborEnvVarsHidden(dockerService *config.ServiceConfig, hiddenEnvVarFileName string) []EnvVarPayload {

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
	if Verbose {
		log.Printf("Looking for hidden environment variables in %s \n", hiddenEnvVarFileName)
	}
	hiddenEnvVars := false
	hiddenEnvVarFile := ""
	for _, envFileName := range dockerService.EnvFile {
		if strings.HasSuffix(envFileName, hiddenEnvVarFileName) {
			hiddenEnvVars = true
			hiddenEnvVarFile = envFileName
			if Verbose {
				log.Println("Found hidden environment variable file: " + hiddenEnvVarFile)
			}
			break
		}
	}

	//iterate/process hidden envvars and remove them from the list
	if hiddenEnvVars {
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
