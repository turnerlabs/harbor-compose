package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// DeserializeHarborCompose deserializes a harbor-compose.yml file into an object
func DeserializeHarborCompose(file string) HarborCompose {

	//read the harbor compose file
	harborComposeData, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	return unmarshalHarborCompose(string(harborComposeData))
}

func marshalHarborCompose(o HarborCompose) []byte {
	//serialize object to yaml
	data, err := yaml.Marshal(o)
	check(err)
	return data
}

// SerializeHarborCompose serializes an object to a harbor-compose.yml file
func SerializeHarborCompose(harborCompose HarborCompose, file string) {

	data := marshalHarborCompose(harborCompose)

	if Verbose {
		log.Printf("writing harbor-compose file to %v", file)
	}

	//write yaml to harbor-compose.yml
	err := ioutil.WriteFile(file, data, 0644)
	if err != nil {
		log.Fatalf("error writing %v: %v", HarborComposeFile, err)
	}

	if Verbose {
		fmt.Println()
		fmt.Printf(string(data))
	}
}
