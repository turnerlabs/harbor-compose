package cmd

import "log"

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
