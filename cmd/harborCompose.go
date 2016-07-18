package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// DeserializeHarborCompose deserializes a harbor-compse.yml file into an object
func DeserializeHarborCompose(file string) HarborCompose {

	//read the harbor compose file
	harborComposeData, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	//parse the harbor compose file
	var harborCompose HarborCompose
	err = yaml.Unmarshal([]byte(harborComposeData), &harborCompose)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return harborCompose
}

// SerializeHarborCompose serializes an object to a harbor-compose.yml file
func SerializeHarborCompose(harborCompose HarborCompose, file string) {

	//serialize object to yaml
	data, err := yaml.Marshal(harborCompose)
	if err != nil {
		log.Fatalf("error marshaling yaml: %v", err)
	}

	if Verbose {
		log.Printf("writing harbor-compose file to %v", file)
	}

	//write yaml to harbor-compose.yml
	err = ioutil.WriteFile(file, data, 0644)
	if err != nil {
		log.Fatalf("error writing %v: %v", HarborComposeFile, err)
	}

	if Verbose {
		fmt.Println()
		fmt.Printf(string(data))
	}
}
