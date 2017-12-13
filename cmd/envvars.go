package cmd

import (
	"strings"
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
