package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func getBuildTokenEnvVar(shipment string, environment string) string {

	//look for envvar for this shipment/environment that matches naming convention: SHIPMENT_ENV_TOKEN
	envvar := fmt.Sprintf("%v_%v_TOKEN", strings.Replace(strings.ToUpper(shipment), "-", "_", -1), strings.ToUpper(environment))
	if Verbose {
		log.Printf("looking for environment variable named: %v\n", envvar)
	}
	buildTokenEnvVar := os.Getenv(envvar)

	//validate build token
	if len(buildTokenEnvVar) == 0 {
		log.Fatalf("A shipment/environment build token is required. Please specify an environment variable named, %v", envvar)
	}

	return buildTokenEnvVar
}
