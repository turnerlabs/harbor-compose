package cmd

import "strings"

func specialEnvVars() map[string]string {
	return map[string]string{
		"CUSTOMER": "CUSTOMER",
		"PRODUCT":  "PRODUCT",
		"PROJECT":  "PROJECT",
		"PROPERTY": "PROPERTY",
		"BARGE":    "BARGE",
	}
}

func copyEnvVars(source []EnvVarPayload, destination map[string]string, special map[string]string) {
	//filter out special metadata envvars and return them
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
