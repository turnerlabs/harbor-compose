package cmd

import (
	"fmt"
	"log"
	"strings"

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

var envvarsShipment string
var envvarsEnvironment string
var envvvarsHiddenFile string
var envvvarsEnvFile string

func init() {
	RootCmd.AddCommand(envvarsCmd)

	//push
	envvarsCmd.AddCommand(pushEnvvarsCmd)
	pushEnvvarsCmd.PersistentFlags().StringVarP(&envvarsShipment, "shipment", "s", "", "shipment name")
	pushEnvvarsCmd.PersistentFlags().StringVarP(&envvarsEnvironment, "environment", "e", "", "environment name")
	pushEnvvarsCmd.PersistentFlags().StringVarP(&envvvarsHiddenFile, "hidden", "", hiddenEnvFileName, "The location of the docker compose environment file that contains hidden environment variables")
}

// envvarsCmd represents the envvars command
var envvarsCmd = &cobra.Command{
	Use:   "envvars",
	Short: "manage environment variables",
	Long:  "manage environment variables",
	Example: `harbor-compose envvars push 
harbor-compose envvars pull`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	PreRun: preRunHook,
}

// pushEnvvarsCmd represents the envvars push command
var pushEnvvarsCmd = &cobra.Command{
	Use:   "push",
	Short: "push docker compose environment variables to harbor",
	Long: `push docker compose environment variables to harbor

The push command works with a harbor-compose.yml file to push environment variables for one or many shipment/environment/containers, as well as for a single shipment environment using the --shipment and --environment flags.
`,
	Example: `harbor-compose push
harbor-compose push -s my-shipment -e dev

You can specify which env file contains your hidden environment variables using the --hidden flag (defaults to hidden.env)
harbor-compose push --hidden my-hidden.env
`,
	Run: pushEnvVars,
}

func pushEnvVars(cmd *cobra.Command, args []string) {

	//make sure user is authenticated
	username, token, err := Login()
	check(err)

	//determine which shipment/environments user wants to process
	inputShipmentEnvironments, harborComposeConfig := getShipmentEnvironmentsFromInput(envvarsShipment, envvarsEnvironment)

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
			harborEnvVars := transformDockerServiceEnvVarsToHarborEnvVarsHidden(serviceConfig, envvvarsHiddenFile)
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
