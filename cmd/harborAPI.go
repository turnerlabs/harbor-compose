package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jtacoma/uritemplates"
	"github.com/parnurzeal/gorequest"
)

func shipitURI(template string, params ...tuple) string {
	return buildURI(GetConfig().ShipitURI, template, params...)
}

func helmitURI(template string, params ...tuple) string {
	return buildURI(GetConfig().HelmitURI, template, params...)
}

func triggerURI(template string, params ...tuple) string {
	return buildURI(GetConfig().TriggerURI, template, params...)
}

func customsURI(template string, params ...tuple) string {
	return buildURI(GetConfig().CustomsURI, template, params...)
}

func bargesURI(template string, params ...tuple) string {
	return buildURI(GetConfig().BargesURI, template, params...)
}

// GetShipmentEnvironment returns a harbor shipment from the API
func GetShipmentEnvironment(username string, token string, shipment string, env string) *ShipmentEnvironment {

	//build URI
	uri := shipitURI("/v1/shipment/{shipment}/environment/{env}/",
		param("shipment", shipment),
		param("env", env))

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
		check(err[0])
	}

	//return nil if the shipment/env isn't found
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatal("GetShipment returned ", resp.StatusCode)
	}

	//deserialize json into object
	var result ShipmentEnvironment
	unmarshalErr := json.Unmarshal(body, &result)
	check(unmarshalErr)

	return &result
}

//UpdateProvider updates provider configuration
func UpdateProvider(username string, token string, shipment string, env string, provider ProviderPayload) {

	uri := shipitURI("/v1/shipment/{shipment}/environment/{env}/provider/ec2",
		param("shipment", shipment),
		param("env", env))

	if Verbose {
		log.Printf("updating replicas on shipment provider: " + uri)
	}

	//call the API
	r, _, e := update(username, token, uri, provider)
	if e != nil {
		check(e[0])
	}
	if r.StatusCode != http.StatusOK {
		check(errors.New("update provider failed"))
	}
}

//UpdateShipmentEnvironment updates shipment/environment-level configuration
func UpdateShipmentEnvironment(username string, token string, shipment string, composeShipment ComposeShipment) {

	//update enableMonitoring
	uri := shipitURI("/v1/shipment/{shipment}/environment/{env}",
		param("shipment", shipment),
		param("env", composeShipment.Env))

	if Verbose {
		log.Printf("updating enableMonitoring on shipment provider: " + uri)
	}

	request := UpdateShipmentEnvironmentRequest{
		EnableMonitoring: *composeShipment.EnableMonitoring,
	}

	//call the API
	r, _, e := update(username, token, uri, request)
	if e != nil {
		check(e[0])
	}
	if r.StatusCode != http.StatusOK {
		check(errors.New("update provider failed"))
	}
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
		check(err[0])
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
		if b, e := json.Marshal(data); e == nil {
			log.Println(string(b))
		}
	}

	res, body, err := gorequest.New().
		Put(url).
		Set("x-username", username).
		Set("x-token", token).
		Send(data).
		End()

	if err != nil {
		check(err[0])
	}

	if Verbose {
		log.Println(body)
		log.Printf("status code = %v", res.StatusCode)
	}

	return res, body, err
}

func deleteHTTP(username string, token string, url string) (*http.Response, string, []error) {

	if Verbose {
		log.Printf("DELETE %v", url)
	}

	res, body, err := gorequest.New().
		Delete(url).
		Set("x-username", username).
		Set("x-token", token).
		End()

	if err != nil {
		check(err[0])
	}

	if Verbose {
		log.Printf("status code = %v", res.StatusCode)
		log.Println(body)
	}

	return res, body, err
}

// GetLogs returns a string of all container logs for a shipment
func GetLogs(barge string, shipment string, env string) string {

	uri := helmitURI("/harbor/{barge}/{shipment}/{env}",
		param("barge", barge),
		param("shipment", shipment),
		param("env", env))

	_, body, err := gorequest.New().
		Get(uri).
		End()

	if err != nil {
		check(err[0])
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

// GetShipmentEvents returns a ShipmentEventResult for a given shipment/environment
func GetShipmentEvents(barge string, shipment string, env string) *ShipmentEventResult {

	uri := helmitURI("/shipment/events/{barge}/{shipment}/{env}",
		param("barge", barge),
		param("shipment", shipment),
		param("env", env))

	if Verbose {
		fmt.Println("fetching: " + uri)
	}

	res, body, err := gorequest.New().
		Get(uri).
		EndBytes()

	if err != nil {
		check(err[0])
	}

	if res.StatusCode != http.StatusOK {
		log.Fatal("GetShipmentEvents returned ", res.StatusCode)
	}

	//deserialize json into object
	var result ShipmentEventResult
	unmarshalErr := json.Unmarshal(body, &result)
	check(unmarshalErr)

	return &result
}

// GetShipmentStatus returns the running status of a shipment
func GetShipmentStatus(barge string, shipment string, env string) *ShipmentStatus {

	uri := helmitURI("/shipment/status/{barge}/{shipment}/{env}",
		param("barge", barge),
		param("shipment", shipment),
		param("env", env))

	if Verbose {
		fmt.Println("fetching: " + uri)
	}

	res, body, err := gorequest.New().
		Get(uri).
		EndBytes()

	if err != nil {
		check(err[0])
	}

	if res.StatusCode != http.StatusOK {
		log.Fatal("GetShipmentStatus returned ", res.StatusCode)
	}

	//deserialize json into object
	var result ShipmentStatus
	unmarshalErr := json.Unmarshal(body, &result)
	check(unmarshalErr)

	return &result
}

// Trigger calls the trigger api
func Trigger(shipment string, env string) (bool, []string) {

	//build URI
	uri := triggerURI("/{shipment}/{env}/ec2",
		param("shipment", shipment),
		param("env", env))

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
		check(err[0])
	}

	//if verbose or non-OK, log status code and message body
	if Verbose || resp.StatusCode != http.StatusOK {
		log.Printf("trigger api returned a %v", resp.StatusCode)
		log.Println(string(body))
	}

	var result []string

	//parse http OK responses as JSON
	if resp.StatusCode == http.StatusOK {
		//trigger api returns both single and multiple messages:

		//example responses...
		//error: {"message":"Could not parse docker image data from http://registry.services.dmtio.net/v2/mss-poc-thingproxy/manifests/: 757: unexpected token at '404 page not found\n'\n"}
		//success: {"message":["compose-test.dev.services.ec2.dmtio.net:5000"]}

		//single message
		if strings.Contains(string(body), "message\":\"") {
			//convert single message into an array for consistency
			var response TriggerResponseSingle
			unmarshalErr := json.Unmarshal(body, &response)
			check(unmarshalErr)
			result = append(result, response.Message)
		} else if strings.Contains(string(body), "message\":[") {
			//multiple messages
			var response TriggerResponseMultiple
			unmarshalErr := json.Unmarshal(body, &response)
			check(unmarshalErr)
			result = response.Messages
		}
	}

	//return whether trigger call was successful along with messages
	return resp.StatusCode == http.StatusOK, result
}

// SaveEnvVar updates an environment variable in harbor (supports both environment and container levels)
func SaveEnvVar(username string, token string, shipment string, environment string, envVarPayload EnvVarPayload, container string) {
	var config = GetConfig()

	//first, issue a GET to check if the var exists
	//if not exists, issue a POST
	//if exists and value has changed, issue a PUT

	//build url
	templateStringEnvLevelExisting := config.ShipitURI + "/v1/shipment/{shipment}/environment/{env}/envvar/{envvar}"
	templateStringContainerLevelExisting := config.ShipitURI + "/v1/shipment/{shipment}/environment/{env}/container/{container}/envvar/{envvar}"
	templateStringEnvLevelNew := config.ShipitURI + "/v1/shipment/{shipment}/environment/{env}/envvars/"
	templateStringContainerLevelNew := config.ShipitURI + "/v1/shipment/{shipment}/environment/{env}/container/{container}/envvars/"

	values := make(map[string]interface{})
	values["shipment"] = shipment
	values["env"] = environment
	values["envvar"] = envVarPayload.Name
	values["container"] = container

	//is the var at the environment or container level?
	templateString := templateStringEnvLevelExisting
	if len(container) > 0 {
		templateString = templateStringContainerLevelExisting
	}
	template, _ := uritemplates.Parse(templateString)
	uri, _ := template.Expand(values)

	//issue GET request
	request := gorequest.New().Get(uri).
		Set("x-username", username).
		Set("x-token", token)

	if Verbose {
		fmt.Println("fetching: " + uri)
	}
	res, body, err := request.EndBytes()
	if err != nil {
		check(err[0])
	}

	//exist?
	if res.StatusCode == http.StatusNotFound { //not exist
		//build url
		//now POST a new envvar
		//is the var at the environment or container level?
		templateString := templateStringEnvLevelNew
		if len(container) > 0 {
			templateString = templateStringContainerLevelNew
		}
		template, _ = uritemplates.Parse(templateString)
		uri, _ = template.Expand(values)

		if Verbose {
			fmt.Println("creating env var...")
		}

		//call the api
		r, _, e := create(username, token, uri, envVarPayload)
		if e != nil {
			check(e[0])
		}
		if r.StatusCode != http.StatusCreated {
			check(errors.New("unable to create env var"))
		}

	} else { //exist
		//issue PUT if modified

		//deserialize json into object
		var result EnvVarPayload
		unmarshalErr := json.Unmarshal(body, &result)
		check(unmarshalErr)

		//modified?
		if result.Value != envVarPayload.Value || result.Type != envVarPayload.Type {
			if Verbose {
				fmt.Println("updating env var...")
			}

			//is the var at the environment or container level?
			templateString := templateStringEnvLevelExisting
			if len(container) > 0 {
				templateString = templateStringContainerLevelExisting
			}
			template, _ = uritemplates.Parse(templateString)
			uri, _ = template.Expand(values)
			r, _, e := update(username, token, uri, envVarPayload)
			if e != nil {
				check(e[0])
			}
			if r.StatusCode != http.StatusOK {
				check(errors.New("unable to update env var"))
			}

		} else {
			if Verbose {
				fmt.Println("envvar unchanged, skipping")
			}
		}
	}
}

// UpdateContainerImage updates a container version on a shipment
func UpdateContainerImage(username string, token string, shipment string, env string, container ContainerPayload) {

	//build url
	uri := shipitURI("/v1/shipment/{shipment}/environment/{env}/container/{container}",
		param("shipment", shipment),
		param("env", env),
		param("container", container.Name))

	if Verbose {
		log.Printf("updating container settings")
	}

	//call api
	r, _, err := update(username, token, uri, container)
	if err != nil {
		check(err[0])
	}
	if r.StatusCode != http.StatusOK {
		check(errors.New("update failed"))
	}
}

// SaveNewShipmentEnvironment bulk saves a new shipment/environment
func SaveNewShipmentEnvironment(username string, token string, shipment ShipmentEnvironment) bool {

	var config = GetConfig()
	shipment.Username = username
	shipment.Token = token

	//POST /api/v1/shipments
	res, body, err := create(username, token, config.ShipitURI+"/v1/bulk/shipments", shipment)

	if err != nil || res.StatusCode != http.StatusCreated {
		fmt.Printf("creating shipment was not successful: %v\ncode: %v\n", body, res.StatusCode)
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
	uri := shipitURI("/v1/shipment/{shipment}/environment/{env}",
		param("shipment", shipment),
		param("env", env))

	if Verbose {
		log.Printf("deleting: " + uri)
	}

	res, _, _ := deleteHTTP(username, token, uri)

	if res.StatusCode != http.StatusOK {
		log.Fatalf("delete returned a status code of %v", res.StatusCode)
	}
}

// Catalogit sends a POST to the catalogit api
func Catalogit(container CatalogitContainer) (string, []error) {

	var config = GetConfig()

	if Verbose {
		log.Printf("Sending POST to: %v /v1/containers\n", config.CatalogitURI)
	}

	//make network request
	resp, body, err := gorequest.New().
		Post(config.CatalogitURI + "/v1/containers").
		Send(container).
		EndBytes()

	//treat non-OK as error
	if resp.StatusCode != http.StatusOK {
		err = append(err, fmt.Errorf("catalogit api returned a %v", resp.StatusCode))
	}

	return string(body), err
}

//IsContainerVersionCataloged determines whether or not a container/version exists in the catalog
func IsContainerVersionCataloged(name string, version string) bool {

	//build URI
	uri := customsURI("/catalog/{name}/{version}/",
		param("name", name),
		param("version", version))

	if Verbose {
		log.Println("fetching: " + uri)
	}

	//issue request
	res, _, err := gorequest.New().Get(uri).EndBytes()
	if err != nil {
		check(err[0])
	}

	//not found
	if res.StatusCode == 404 {
		return false
	}

	//throw error if not OK
	if res.StatusCode != http.StatusOK {
		log.Fatalf("GET %v returned %v", uri, res.StatusCode)
	}

	return true
}

// Deploy deploys (and catalogs) a shipment container to an environment
func Deploy(shipment string, env string, buildToken string, deployRequest DeployRequest, provider string) {

	//build URI
	uri := customsURI("/deploy/{shipment}/{env}/{provider}",
		param("shipment", shipment),
		param("env", env),
		param("provider", provider))

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
		check(err[0])
	}

	//logging
	if Verbose || res.StatusCode != http.StatusOK {
		log.Printf("customs api returned a %v", res.StatusCode)
		log.Println(string(body))
	}

	if res.StatusCode != http.StatusOK {
		check(errors.New("customs/deploy failed"))
	}
}

// CatalogCustoms catalogs a container using the customs catalog api
func CatalogCustoms(shipment string, env string, buildToken string, catalogRequest CatalogitContainer, provider string) {

	uri := customsURI("/catalog/{shipment}/{env}/{provider}",
		param("shipment", shipment),
		param("env", env),
		param("provider", provider))

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
		check(err[0])
	}

	//logging
	if Verbose || res.StatusCode != http.StatusOK {
		log.Printf("customs api returned a %v", res.StatusCode)
		log.Println(string(body))
	}

	if res.StatusCode != http.StatusOK {
		check(errors.New("customs/catalog failed"))
	}
}

//update a port
func updatePort(username string, token string, shipment string, env string, container string, port UpdatePortRequest) {

	//build url
	uri := shipitURI("/v1/shipment/{shipment}/environment/{env}/container/{container}/port/{port}",
		param("shipment", shipment),
		param("env", env),
		param("container", container),
		param("port", port.Name))

	//make the api call
	r, _, e := update(username, token, uri, port)
	if e != nil {
		check(e[0])
	}
	if r.StatusCode != http.StatusOK {
		check(errors.New("update port failed"))
	}
}

// GetBarges returns a list of harbor barges
func GetBarges() *BargeResults {

	uri := bargesURI("/barges")
	if Verbose {
		fmt.Println("fetching: " + uri)
	}

	res, body, err := gorequest.New().Get(uri).EndBytes()
	if err != nil {
		check(err[0])
	}

	if res.StatusCode != http.StatusOK {
		log.Fatal("GetBarges returned ", res.StatusCode)
	}

	//deserialize json into object
	var result BargeResults
	unmarshalErr := json.Unmarshal(body, &result)
	check(unmarshalErr)

	return &result
}

// GetGroup returns the members of a harbor group
func GetGroup(id string) *Group {

	uri := bargesURI("/harbor/groups/{id}", param("id", id))

	if Verbose {
		fmt.Println("fetching: " + uri)
	}

	res, body, err := gorequest.New().Get(uri).EndBytes()
	if err != nil {
		check(err[0])
	}

	if res.StatusCode != http.StatusOK {
		log.Fatal("GetBarges returned ", res.StatusCode)
	}

	//deserialize json into object
	var result Group
	unmarshalErr := json.Unmarshal(body, &result)
	check(unmarshalErr)

	return &result
}
