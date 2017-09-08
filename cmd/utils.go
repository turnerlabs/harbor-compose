package cmd

import (
	"io/ioutil"
	"log"
	"os"
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

//find the ec2 provider
func ec2ProviderNewProvider(providers []NewProvider) *NewProvider {
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
