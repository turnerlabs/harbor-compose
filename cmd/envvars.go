package cmd

import "strings"

const (
	envVarNameCustomer = "CUSTOMER"
	envVarNameProduct  = "PRODUCT"
	envVarNameProject  = "PROJECT"
	envVarNameProperty = "PROPERTY"
	envVarNameBarge    = "BARGE"
	envVarNameRestart  = "HC_RESTART"
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

func copyEnvVars(source []EnvVarPayload, destination map[string]string, special map[string]string) {
	//filter out special envvars and return them
	for _, envvar := range source {
		if specialEnvVars()[strings.ToUpper(envvar.Name)] == "" {
			if destination != nil {
				destination[envvar.Name] = envvar.Value
			}
		} else if special != nil {
			special[envvar.Name] = envvar.Value
		}
	}
}
