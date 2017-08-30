package cmd

import "strings"

func specialEnvVars() map[string]string {
	return map[string]string{
		"CUSTOMER":   "CUSTOMER",
		"PRODUCT":    "PRODUCT",
		"PROJECT":    "PROJECT",
		"PROPERTY":   "PROPERTY",
		"BARGE":      "BARGE",
		"HC_RESTART": "HC_RESTART",
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
