package cmd

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/jtacoma/uritemplates"
)

func check(e error) {
	if e != nil {
		log.Fatal("ERROR: ", e)
	}
}

//find the ec2 provider
func ec2Provider(providers []ProviderPayload) *ProviderPayload {
	for _, provider := range providers {
		if provider.Name == providerEc2 {
			return &provider
		}
	}
	log.Fatal("ec2 provider is missing")
	return nil
}

func appendToFile(file string, lines []string) {
	if _, err := os.Stat(file); err == nil {
		//update
		file, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
		check(err)
		defer file.Close()
		for _, line := range lines {
			_, err = file.WriteString("\n" + line)
			check(err)
		}
	} else {
		//create
		data := ""
		for _, line := range lines {
			data += line + "\n"
		}
		err := ioutil.WriteFile(file, []byte(data), 0644)
		check(err)
	}
}

type tuple struct {
	Item1 string
	Item2 string
}

func param(item1 string, item2 string) tuple {
	return tuple{
		Item1: item1,
		Item2: item2,
	}
}

func buildURI(baseURI string, template string, params ...tuple) string {
	uriTemplate, err := uritemplates.Parse(baseURI + template)
	check(err)
	values := make(map[string]interface{})
	for _, v := range params {
		values[v.Item1] = v.Item2
	}
	uri, err := uriTemplate.Expand(values)
	check(err)
	return uri
}

func findContainer(container string, containers []ContainerPayload) *ContainerPayload {
	for _, c := range containers {
		if c.Name == container {
			return &c
		}
	}
	return nil
}
