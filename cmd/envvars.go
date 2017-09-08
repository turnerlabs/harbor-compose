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

//processes envvars by copying them to a destination and filtering out special and hidden envvars
func copyEnvVars(source []EnvVarPayload, destination map[string]string, special map[string]string, hidden map[string]string) {

	//iterate all envvars
	for _, envvar := range source {

		//is this a special envvar?
		if specialEnvVars()[strings.ToUpper(envvar.Name)] == "" { //no

			//hidden?
			if envvar.Type == "hidden" && hidden != nil {
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
