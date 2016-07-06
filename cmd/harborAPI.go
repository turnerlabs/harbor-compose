package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jtacoma/uritemplates"
	"github.com/parnurzeal/gorequest"
)

var shipItURI = "http://shipit.services.dmtio.net"
var triggerURI = "http://harbor-trigger.services.dmtio.net"
var authAPI = "http://auth.services.dmtio.net"
var helmitURI = "http://helmit.services.dmtio.net"

// GetShipmentEnvironment returns a harbor shipment from the API
func GetShipmentEnvironment(shipment string, env string) *ShipmentEnvironment {

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
	resp, body, err := gorequest.New().
		Get(uri).
		EndBytes()

	if err != nil {
		log.Fatal(err)
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

// GetToken returns a token
func GetToken(user string, passwd string) string {

	var url = authAPI + "/v1/auth/gettoken"

	var data = AuthRequest{
		User: user,
		Pass: passwd,
	}

	_, body, err := gorequest.New().
		Post(url).
		Send(data).
		EndBytes()

	if err != nil {
		log.Fatal(err)
	}

	var authResponse AuthResponse
	unErr := json.Unmarshal(body, &authResponse)
	if unErr != nil {
		log.Fatal(unErr)
	}

	var token string

	if authResponse.Success {
		if Verbose {
			log.Printf("User %s has been authenticated", user)
		}
		token = authResponse.Token
	} else {
		log.Fatalf("There was an error authenticating %s", user)
	}

	return token
}

//UpdateShipment updates shipment-level configuration
func UpdateShipment(shipment string, composeShipment ComposeShipment, token string) {

	//PUT /v1/shipment/:Shipment/environment/:Environment/provider/:name
	//build URI
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
	update(token, uri, providerPayload)
}

func update(token string, url string, data interface{}) {

	_, body, err := gorequest.New().
		Put(url).
		Set("x-username", User).
		Set("x-token", token).
		Send(data).
		End()

	if err != nil {
		log.Fatal(err)
	}

	if Verbose {
		log.Println(body)
	}
}

// GetLogs returns a string of all container logs for a shipment
func GetLogs(barge string, shipment string, env string) string {
    var url string = helmitURI + "/harbor/" + barge + "/" + shipment + "/" + env

		log.Println(url)
		_, body, err := gorequest.New().
			Get(url).
			End()

		if err != nil {
			log.Fatal(err)
		}

		if Verbose {
		  fmt.Println("Fetching Harbor Logs")
		}

		return body
}

// Trigger calls the trigger api
func Trigger(shipment string, env string) {

	//build URI
	values := make(map[string]interface{})
	values["shipment"] = shipment
	values["env"] = env
	template, _ := uritemplates.Parse(triggerURI + "/{shipment}/{env}/ec2")
	uri, _ := template.Expand(values)
	if Verbose {
		log.Printf("triggering shipment: " + uri)
	}

	_, body, err := gorequest.New().
		Post(uri).
		End()

	if err != nil {
		log.Fatal(err)
	}

	if Verbose {
		log.Println(body)
	}
}
