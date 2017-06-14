package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
)

// Config is the config for all communications in harbor
type Config struct {
	ShipitURI    string `json:"shipit"`
	CatalogitURI string `json:"catalogit"`
	TriggerURI   string `json:"trigger"`
	AuthURI      string `json:"authn"`
	HelmitURI    string `json:"helmit"`
	HarborURI    string `json:"harbor"`
	CustomsURI   string `json:"customs"`
}

func readConfig() (*Config, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	curpath, _ := os.Getwd()

	var hcConfig = os.Getenv("HC_CONFIG")

	var path = home + "/.harbor"
	err = os.Chdir(path)
	if err != nil {
		err = os.Mkdir(path, 0700)
		if err != nil {
			return nil, err
		}
	}

	var configPath = path + "/config"

	if hcConfig != "" {
		configPath = hcConfig
	}

	if Verbose {
		log.Println(configPath)
	}

	var serializedConfig Config
	byteData, err := ioutil.ReadFile(configPath)
	_ = os.Chdir(curpath)
	if err != nil || isJSON(string(byteData)) == false {
		serializedConfig := new(Config)
		return serializedConfig, nil
	}

	err = json.Unmarshal(byteData, &serializedConfig)
	if err != nil {
		return nil, err
	}

	return &serializedConfig, nil
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil

}

// GetConfig will set the default values
func GetConfig() *Config {

	var config, err = readConfig()

	if err != nil {
		log.Fatal(err)
	}

	if config.ShipitURI == "" {
		config.ShipitURI = "http://shipit.services.dmtio.net"
	}

	if config.CatalogitURI == "" {
		config.CatalogitURI = "http://catalogit.services.dmtio.net"
	}

	if config.TriggerURI == "" {
		config.TriggerURI = "http://harbor-trigger.services.dmtio.net"
	}

	if config.AuthURI == "" {
		config.AuthURI = "https://auth.services.dmtio.net"
	}

	if config.HelmitURI == "" {
		config.HelmitURI = "http://helmit.services.dmtio.net"
	}

	if config.HarborURI == "" {
		config.HarborURI = "http://harbor.services.dmtio.net"
	}

	if config.CustomsURI == "" {
		config.CustomsURI = "https://customs.services.dmtio.net"
	}

	return config
}
