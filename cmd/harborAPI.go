package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jtacoma/uritemplates"
	"github.com/parnurzeal/gorequest"
)

var shipItURI = "http://shipit.services.dmtio.net"
var triggerURI = "http://harbor-trigger.services.dmtio.net"
var authAPI = "http://auth.services.dmtio.net"
var helmitURI = "http://helmit.services.dmtio.net"
var harborURI = "http://harbor.services.dmtio.net"

// GetShipmentEnvironment returns a harbor shipment from the API
func GetShipmentEnvironment(shipment string, env string) *ShipmentEnvironment {
	userName, token, _ := Login()

	//build URI
	values := make(map[string]interface{})
	values["shipment"] = shipment
	values["env"] = env
	template, _ := uritemplates.Parse(shipItURI + "/v1/shipment/{shipment}/environment/{env}/")
	uri, _ := template.Expand(values)
	if Verbose {
		fmt.Println("fetching: " + uri)
	}

	//issue request
	request := gorequest.New().Get(uri)

	//if token is specified, add it to the headers
	if token != "" {
		request = request.
			Set("x-username", userName).
			Set("x-token", token)
	}

	resp, body, err := request.EndBytes()

	if err != nil {
		log.Fatal(err)
	}

	//return nil if the shipment/env isn't found
	if resp.StatusCode == 404 {
		return nil
	}

	if resp.StatusCode != 200 {
		log.Fatal("GetShipment returned ", resp.StatusCode)
	}

	//deserialize json into object
	var result ShipmentEnvironment
	unmarshalErr := json.Unmarshal(body, &result)
	if unmarshalErr != nil {
		log.Fatal(unmarshalErr)
	}

	return &result
}

//UpdateShipment updates shipment-level configuration
func UpdateShipment(shipment string, composeShipment ComposeShipment) {
	//build URI
	//PUT /v1/shipment/:Shipment/environment/:Environment/provider/:name
	values := make(map[string]interface{})
	values["shipment"] = shipment
	values["env"] = composeShipment.Env
	template, _ := uritemplates.Parse(shipItURI + "/v1/shipment/{shipment}/environment/{env}/provider/ec2")
	uri, _ := template.Expand(values)
	if Verbose {
		log.Printf("updating replicas on shipment provider: " + uri)
	}

	providerPayload := ProviderPayload{
		Name:     "ec2",
		Replicas: composeShipment.Replicas,
	}

	//call the API
	update(uri, providerPayload)
}

func create(url string, data interface{}) (*http.Response, string, []error) {
	userName, token, _ := Login()
	res, body, err := gorequest.New().
		Post(url).
		Set("x-username", userName).
		Set("x-token", token).
		Send(data).
		End()

	if err != nil {
		log.Fatal(err)
	}

	return res, body, err
}

func update(url string, data interface{}) (*http.Response, string, []error) {
	userName, token, _ := Login()

	res, body, err := gorequest.New().
		Put(url).
		Set("x-username", userName).
		Set("x-token", token).
		Send(data).
		End()

	if err != nil {
		log.Fatal(err)
	}

	if Verbose {
		log.Println(body)
	}

	return res, body, err
}

func delete(url string) (*http.Response, string, []error) {
	userName, token, _ := Login()

	res, body, err := gorequest.New().
		Delete(url).
		Set("x-username", userName).
		Set("x-token", token).
		End()

	if err != nil {
		log.Fatal(err)
	}

	if Verbose {
		log.Println(body)
	}

	return res, body, err
}

// GetLogs returns a string of all container logs for a shipment
func GetLogs(barge string, shipment string, env string) string {
	values := make(map[string]interface{})
	values["barge"] = barge
	values["shipment"] = shipment
	values["env"] = env
	template, _ := uritemplates.Parse(helmitURI + "/harbor/{barge}/{shipment}/{env}")
	uri, _ := template.Expand(values)

	_, body, err := gorequest.New().
		Get(uri).
		End()

	if err != nil {
		log.Fatal(err)
	}

	if Verbose {
		fmt.Println(uri)
		fmt.Println("Fetching Harbor Logs")
	}

	return body
}

// Trigger calls the trigger api
func Trigger(shipment string, env string) (bool, []string) {

	//build URI
	values := make(map[string]interface{})
	values["shipment"] = shipment
	values["env"] = env
	template, _ := uritemplates.Parse(triggerURI + "/{shipment}/{env}/ec2")
	uri, _ := template.Expand(values)
	if Verbose {
		log.Printf("triggering shipment: " + uri)
	}

	resp, body, err := gorequest.New().
		Post(uri).
		EndBytes()

	if err != nil {
		log.Fatal(err)
	}

	//example responses...
	//error: {"message":"Could not parse docker image data from http://registry.services.dmtio.net/v2/mss-poc-thingproxy/manifests/: 757: unexpected token at '404 page not found\n'\n"}
	//success: {"message":["compose-test.dev.services.ec2.dmtio.net:5000"]}

	//trigger api returns both single and multiple messages
	if strings.Contains(string(body), "message\":\"") {

		var response TriggerResponseSingle
		unmarshalErr := json.Unmarshal(body, &response)
		if unmarshalErr != nil {
			log.Fatal(unmarshalErr)
		}

		//convert single message into an array for consistency
		var temp []string
		temp = append(temp, response.Message)

		//return success
		return resp.StatusCode == 200, temp
	}

	var response TriggerResponseMultiple
	unmarshalErr := json.Unmarshal(body, &response)
	if unmarshalErr != nil {
		log.Fatal(unmarshalErr)
	}

	return resp.StatusCode == 200, response.Messages
}

// SaveEnvVar saves envvars by doing a delete/add against the api
func SaveEnvVar(shipment string, composeShipment ComposeShipment, envVarPayload EnvVarPayload, container string) {

	templateString := shipItURI + "/v1/shipment/{shipment}/environment/{env}/envvar/{envvar}"

	//build url
	//DELETE /v1/shipment/%s/environment/%s/envVar
	values := make(map[string]interface{})
	values["shipment"] = shipment
	values["env"] = composeShipment.Env
	values["envvar"] = envVarPayload.Name

	if len(container) > 0 {
		values["container"] = container
		templateString = shipItURI + "/v1/shipment/{shipment}/environment/{env}/container/{container}/envvar/{envvar}"
	}

	template, _ := uritemplates.Parse(templateString)
	url, _ := template.Expand(values)

	//issue delete call
	//api will return 422 if the envvar doesn't exist, which can be ignored
	res, _, _ := delete(url)

	//throw an error if we don't get our expected status code
	if !(res.StatusCode == 200 || res.StatusCode == 422) {
		log.Fatalf("DELETE %v returned %v", url, res.StatusCode)
	}

	//build url
	//now POST a new envvar
	templateString = shipItURI + "/v1/shipment/{shipment}/environment/{env}/envvars"
	if len(container) > 0 {
		values["container"] = container
		templateString = shipItURI + "/v1/shipment/{shipment}/environment/{env}/container/{container}/envvars"
	}
	template, _ = uritemplates.Parse(templateString)
	url, _ = template.Expand(values)

	//call the api
	create(url, envVarPayload)
}

// UpdateContainerImage updates a container version on a shipment
func UpdateContainerImage(shipment string, composeShipment ComposeShipment, container string, dockerService DockerComposeService) {
	if Verbose {
		log.Printf("updating container settings")
	}

	//build url
	//PUT /v1/shipment/%s/environment/%s/container/%s
	values := make(map[string]interface{})
	values["shipment"] = shipment
	values["env"] = composeShipment.Env
	values["container"] = container
	template, _ := uritemplates.Parse(shipItURI + "/v1/shipment/{shipment}/environment/{env}/container/{container}")
	url, _ := template.Expand(values)

	var payload = ContainerPayload{
		Name:  container,
		Image: dockerService.Image,
	}

	//call api
	update(url, payload)
}

// SaveNewShipmentEnvironment bulk saves a new shipment/environment
func SaveNewShipmentEnvironment(shipment NewShipmentEnvironment) bool {
	//POST /api/v1/shipments
	res, body, err := create(harborURI+"/api/v1/shipments", shipment)

	if err != nil || res.StatusCode != 200 {
		fmt.Printf("creating shipment was not successful: %v \n", body)
		return false
	}

	//api returns an object with an errors property that is
	//false when there are no errors and an object if there are
	if !strings.Contains(body, "errors\": false") {
		return false
	}

	return true
}

// DeleteShipmentEnvironment deletes a shipment/environment from harbor
func DeleteShipmentEnvironment(shipment string, env string) {
	//build URI
	values := make(map[string]interface{})
	values["shipment"] = shipment
	values["env"] = env
	template, _ := uritemplates.Parse(shipItURI + "/v1/shipment/{shipment}/environment/{env}")
	uri, _ := template.Expand(values)
	if Verbose {
		log.Printf("deleting: " + uri)
	}

	res, _, _ := delete(uri)

	if res.StatusCode != 200 {
		log.Fatalf("delete returned a status code of %v", res.StatusCode)
	}
}
