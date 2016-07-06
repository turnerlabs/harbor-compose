package cmd

import (
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
