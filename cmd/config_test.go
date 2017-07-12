package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	config := GetConfig()

	assert.Equal(t, "http://shipit.services.dmtio.net", config.ShipitURI)
	assert.Equal(t, "http://catalogit.services.dmtio.net", config.CatalogitURI)
	assert.Equal(t, "https://customs.services.dmtio.net", config.CustomsURI)
	assert.Equal(t, "http://helmit.services.dmtio.net", config.HelmitURI)
	assert.Equal(t, "http://harbor.services.dmtio.net", config.HarborURI)
	assert.Equal(t, "http://harbor-trigger.services.dmtio.net", config.TriggerURI)
	assert.Equal(t, "https://auth.services.dmtio.net", config.AuthURI)
}

func TestReadHarborEndpointsCustom(t *testing.T) {

	writeConfig()
	config := GetConfig()

	assert.Equal(t, "http://shipit.foo.com", config.ShipitURI)
	assert.Equal(t, "http://catalogit.foo.com", config.CatalogitURI)
	assert.Equal(t, "http://customs.foo.com", config.CustomsURI)
	assert.Equal(t, "http://helmit.foo.com", config.HelmitURI)
	assert.Equal(t, "http://harbor.foo.com", config.HarborURI)
	assert.Equal(t, "http://trigger.foo.com", config.TriggerURI)
	assert.Equal(t, "https://auth.foo.com", config.AuthURI)

	removeConfig()
}

func writeConfig() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	configFile := home + "/.harbor/test-config.json"
	os.Setenv("HC_CONFIG", configFile)

	config := new(Config)
	config.ShipitURI = "http://shipit.foo.com"
	config.CatalogitURI = "http://catalogit.foo.com"
	config.CustomsURI = "http://customs.foo.com"
	config.HelmitURI = "http://helmit.foo.com"
	config.HarborURI = "http://harbor.foo.com"
	config.TriggerURI = "http://trigger.foo.com"
	config.AuthURI = "https://auth.foo.com"

	b, _ := json.Marshal(config)

	configByte := []byte(string(b))
	err = ioutil.WriteFile(configFile, configByte, 0600)
	if err != nil {
		log.Fatal(err)
	}
}

func removeConfig() {
	config := os.Getenv("HC_CONFIG")
	err := os.Remove(config)

	if err != nil {
		log.Fatal(err)
	}
}
