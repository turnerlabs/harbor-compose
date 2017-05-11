package cmd

import (
	"bufio"
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
var catalogitURI = "http://catalogit.services.dmtio.net"
var customsURI = "https://customs.services.dmtio.net"

// GetShipmentEnvironment returns a harbor shipment from the API
func GetShipmentEnvironment(username string, token string, shipment string, env string) *ShipmentEnvironment {
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
			Set("x-username", username).
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
func UpdateShipment(username string, token string, shipment string, composeShipment ComposeShipment) {
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
	update(username, token, uri, providerPayload)
}

func create(username string, token string, url string, data interface{}) (*http.Response, string, []error) {

	if Verbose {
		log.Printf("POST %v", url)
	}

	res, body, err := gorequest.New().
		Post(url).
		Set("x-username", username).
		Set("x-token", token).
		Send(data).
		End()

	if err != nil {
		log.Fatal(err)
	}

	if Verbose {
		log.Printf("status code = %v", res.StatusCode)
		log.Println(body)
	}

	return res, body, err
}

func update(username string, token string, url string, data interface{}) (*http.Response, string, []error) {

	if Verbose {
		log.Printf("PUT %v", url)
	}

	res, body, err := gorequest.New().
		Put(url).
		Set("x-username", username).
		Set("x-token", token).
		Send(data).
		End()

	if err != nil {
		log.Fatal(err)
	}

	if Verbose {
		log.Printf("status code = %v", res.StatusCode)
		log.Println(body)
	}

	return res, body, err
}

func delete(username string, token string, url string) (*http.Response, string, []error) {

	if Verbose {
		log.Printf("DELETE %v", url)
	}

	res, body, err := gorequest.New().
		Delete(url).
		Set("x-username", username).
		Set("x-token", token).
		End()

	if err != nil {
		log.Fatal(err)
	}

	if Verbose {
		log.Printf("status code = %v", res.StatusCode)
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

// GetLogStreamer return reader object to parse docker container logs
func GetLogStreamer(streamer string) (reader *bufio.Reader, err error) {
	resp, err := http.Get(streamer)

	if err != nil {
		return
	}

	reader = bufio.NewReader(resp.Body)
	return
}

// GetShipmentStatus returns the running status of a shipment
func GetShipmentStatus(barge string, shipment string, env string) *ShipmentStatus {

	//build URI
	values := make(map[string]interface{})
	values["barge"] = barge
	values["shipment"] = shipment
	values["env"] = env
	template, _ := uritemplates.Parse(helmitURI + "/shipment/status/{barge}/{shipment}/{env}")
	uri, _ := template.Expand(values)
	if Verbose {
		fmt.Println("fetching: " + uri)
	}

	res, body, err := gorequest.New().
		Get(uri).
		EndBytes()

	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != 200 {
		log.Fatal("GetShipmentStatus returned ", res.StatusCode)
	}

	//deserialize json into object
	var result ShipmentStatus
	unmarshalErr := json.Unmarshal(body, &result)
	if unmarshalErr != nil {
		log.Fatal(unmarshalErr)
	}

	return &result
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

	//make network request
	resp, body, err := gorequest.New().
		Post(uri).
		EndBytes()

	//handle errors
	if err != nil {
		log.Println("an error occurred calling trigger api")
		log.Fatal(err)
	}

	//if verbose or non-200, log status code and message body
	if Verbose || resp.StatusCode != 200 {
		log.Printf("trigger api returned a %v", resp.StatusCode)
		log.Println(string(body))
	}

	var result []string

	//parse http 200 responses as JSON
	if resp.StatusCode == 200 {
		//trigger api returns both single and multiple messages:

		//example responses...
		//error: {"message":"Could not parse docker image data from http://registry.services.dmtio.net/v2/mss-poc-thingproxy/manifests/: 757: unexpected token at '404 page not found\n'\n"}
		//success: {"message":["compose-test.dev.services.ec2.dmtio.net:5000"]}

		//single message
		if strings.Contains(string(body), "message\":\"") {
			//convert single message into an array for consistency
			var response TriggerResponseSingle
			unmarshalErr := json.Unmarshal(body, &response)
			if unmarshalErr != nil {
				log.Fatal(unmarshalErr)
			}
			result = append(result, response.Message)
		} else if strings.Contains(string(body), "message\":[") {
			//multiple messages
			var response TriggerResponseMultiple
			unmarshalErr := json.Unmarshal(body, &response)
			if unmarshalErr != nil {
				log.Fatal(unmarshalErr)
			}
			result = response.Messages
		}
	}

	//return whether trigger call was successful along with messages
	return resp.StatusCode == 200, result
}

// SaveEnvVar saves envvars by doing a delete/add against the api
func SaveEnvVar(username string, token string, shipment string, composeShipment ComposeShipment, envVarPayload EnvVarPayload, container string) {

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
	res, _, _ := delete(username, token, url)

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
	create(username, token, url, envVarPayload)
}

// UpdateContainerImage updates a container version on a shipment
func UpdateContainerImage(username string, token string, shipment string, composeShipment ComposeShipment, container string, image string) {
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
		Image: image,
	}

	//call api
	update(username, token, url, payload)
}

// SaveNewShipmentEnvironment bulk saves a new shipment/environment
func SaveNewShipmentEnvironment(username string, token string, shipment NewShipmentEnvironment) bool {

	shipment.Username = username
	shipment.Token = token

	//POST /api/v1/shipments
	res, body, err := create(username, token, harborURI+"/api/v1/shipments", shipment)

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
func DeleteShipmentEnvironment(username string, token string, shipment string, env string) {
	//build URI
	values := make(map[string]interface{})
	values["shipment"] = shipment
	values["env"] = env
	template, _ := uritemplates.Parse(shipItURI + "/v1/shipment/{shipment}/environment/{env}")
	uri, _ := template.Expand(values)
	if Verbose {
		log.Printf("deleting: " + uri)
	}

	res, _, _ := delete(username, token, uri)

	if res.StatusCode != 200 {
		log.Fatalf("delete returned a status code of %v", res.StatusCode)
	}
}

// Catalogit sends a POST to the catalogit api
func Catalogit(container CatalogitContainer) (response string, err []error) {

	if Verbose {
		log.Printf("Sending POST to: %v /v1/containers\n", catalogitURI)
	}

	//make network request
	resp, body, err := gorequest.New().
		Post(catalogitURI + "/v1/containers").
		Send(container).
		EndBytes()

	// handle errors
	if err != nil && resp.StatusCode != 422 {
		log.Println("an error occurred calling catalogit api")
		log.Fatal(err)
	}

	if Verbose && resp.StatusCode == 422 {
		log.Println("container has already been cataloged.")
	}

	// if verbose or non-200, log status code and message body
	if Verbose || (resp.StatusCode != 200 && resp.StatusCode != 422) {
		log.Printf("catalogit api returned a %v", resp.StatusCode)
		log.Println(string(body))
	}

	return "Successfully Cataloged Container " + container.Name, nil
}

//IsContainerVersionCataloged determines whether or not a container/version exists in the catalog
func IsContainerVersionCataloged(name string, version string) bool {

	//build URI
	values := make(map[string]interface{})
	values["name"] = name
	values["version"] = version
	template, _ := uritemplates.Parse(customsURI + "/catalog/{name}/{version}/")
	uri, _ := template.Expand(values)
	if Verbose {
		log.Println("fetching: " + uri)
	}

	//issue request
	res, _, err := gorequest.New().Get(uri).EndBytes()

	if err != nil {
		log.Fatal(err)
	}

	//not found
	if res.StatusCode == 404 {
		return false
	}

	//throw error if not 200
	if res.StatusCode != 200 {
		log.Fatalf("GET %v returned %v", uri, res.StatusCode)
	}

	return true
}

// Deploy deploys (and catalogs) a shipment container to an environment
func Deploy(shipment string, env string, buildToken string, deployRequest DeployRequest, provider string) {

	//build URI
	values := make(map[string]interface{})
	values["shipment"] = shipment
	values["env"] = env
	values["provider"] = provider
	template, _ := uritemplates.Parse(customsURI + "/deploy/{shipment}/{env}/{provider}")
	uri, _ := template.Expand(values)
	request := gorequest.New().Get(uri)

	if Verbose {
		log.Printf("POST " + uri)
	}

	//make network request
	res, body, err := request.
		Post(uri).
		Set("x-build-token", buildToken).
		Send(deployRequest).
		EndBytes()

	//handle errors
	if err != nil {
		log.Println("an error occurred calling customs api")
		log.Fatal(err)
	}

	//logging
	if Verbose || res.StatusCode != 200 {
		log.Printf("customs api returned a %v", res.StatusCode)
		log.Println(string(body))
	}

	if res.StatusCode != 200 {
		log.Fatal("customs/deploy failed")
	}
}

// CatalogCustoms catalogs a container using the customs catalog api
func CatalogCustoms(shipment string, env string, buildToken string, catalogRequest CatalogitContainer, provider string) {

	//POST /catalog/:shipment/:environment/:provider
	values := make(map[string]interface{})
	values["shipment"] = shipment
	values["env"] = env
	values["provider"] = provider
	template, _ := uritemplates.Parse(customsURI + "/catalog/{shipment}/{env}/{provider}")
	uri, _ := template.Expand(values)
	request := gorequest.New().Get(uri)

	if Verbose {
		log.Printf("POST " + uri)
	}

	//make network request
	res, body, err := request.
		Post(uri).
		Set("x-build-token", buildToken).
		Send(catalogRequest).
		EndBytes()

	//handle errors
	if err != nil {
		log.Println("an error occurred calling customs api")
		log.Fatal(err)
	}

	//logging
	if Verbose || res.StatusCode != 200 {
		log.Printf("customs api returned a %v", res.StatusCode)
		log.Println(string(body))
	}

	if res.StatusCode != 200 {
		log.Fatal("customs/catalog failed")
	}
}
