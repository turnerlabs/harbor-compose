package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getSampleShipmentJSONForLogShipping() string {
	return `
{
  "name": "${env}",
  "parentShipment": {
    "name": "${name}",
    "group": "${group}",
    "envVars": [
      {
        "type": "basic",
        "value": "customer",
        "name": "CUSTOMER"
      },
      {
        "type": "basic",
        "value": "${product}",
        "name": "PRODUCT"
      },
      {
        "type": "basic",
        "value": "${project}",
        "name": "PROJECT"
      },
      {
        "type": "basic",
        "value": "${property}",
        "name": "PROPERTY"
      }
    ]
  },
  "envVars": [
    {      
      "type": "basic",
      "value": "${shipLogs}",
      "name": "SHIP_LOGS"
		},
    {
      "type": "basic",
      "value": "${logsEndpoint}",
      "name": "LOGS_ENDPOINT"
    },
    {
      "type": "basic",
      "value": "${accessKey}",
      "name": "LOGS_ACCESS_KEY"
		},
    {
      "type": "basic",
      "value": "${secretKey}",
      "name": "LOGS_SECRET_KEY"
		},
    {
      "type": "basic",
      "value": "${domainName}",
      "name": "LOGS_DOMAIN_NAME"
		},		
    {
      "type": "basic",
      "value": "${region}",
      "name": "LOGS_REGION"
		},
    {
      "type": "basic",
      "value": "${queueName}",
      "name": "LOGS_QUEUE_NAME"
    }								
  ],
  "providers": [
    {
      "replicas": ${replicas},
      "barge": "${barge}",
      "name": "ec2",
      "envVars": []
    }
  ],
  "containers": [
    {
      "image": "quay.io/turner/web:1.0",
      "name": "${container}",
      "envVars": [
        {
          "type": "basic",
          "value": "/hc",
          "name": "HEALTHCHECK"
        },
        {
          "type": "basic",
          "value": "${containerLevel}",
          "name": "CONTAINER_LEVEL"
				},
				{
					"type": "basic",
					"value": "${containerLevel}",
					"name": "HC_RESTART"
				}				
      ],
      "ports": [
        {
          "protocol": "http",
          "healthcheck": "/hc",
          "external": true,
          "primary": true,
          "public_vip": false,
          "enable_proxy_protocol": false,
          "ssl_arn": "",
          "ssl_management_type": "iam",
          "healthcheck_timeout": 1,
          "public_port": 80,
          "value": 5000,
          "name": "PORT"
        }
      ]
    }
  ]
}	
`
}

func TestGenerateTerraformSourceCodeBasic(t *testing.T) {

	shipmentJSON := getSampleShipmentJSON()

	//update json with test values
	name := "mss-poc-app"
	env := "dev"
	barge := "digital-sandbox"
	replicas := 2
	group := "mss"
	foo := "bar"
	project := "project"
	property := "property"
	product := "product"
	envLevel := "ENV_LEVEL"
	containerLevel := "CONTAINER_LEVEL"
	container := "web"
	healthcheckTimeout := 10
	healthcheckInterval := 100

	shipmentJSON = strings.Replace(shipmentJSON, "${name}", name, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${env}", env, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${barge}", barge, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${replicas}", strconv.Itoa(replicas), 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${group}", group, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${foo}", foo, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${property}", property, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${product}", product, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${project}", project, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${envLevel}", envLevel, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${containerLevel}", containerLevel, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${container}", container, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${healthcheckTimeout}", strconv.Itoa(healthcheckTimeout), 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${healthcheckInterval}", strconv.Itoa(healthcheckInterval), 1)
	t.Log(shipmentJSON)

	//deserialize shipit json
	var shipmentEnv ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipmentEnv)
	if err != nil {
		log.Fatal(err)
	}

	//package the data to make it easy for rendering
	harborCompose := transformShipmentToHarborCompose(&shipmentEnv)
	data := getTerraformData(&shipmentEnv, &harborCompose)

	//generate code
	tf := generateTerraformSourceCode(data)
	t.Log(tf)

	//provider
	assert.Contains(t, tf, "provider \"harbor\"")

	//shipment
	assert.Contains(t, tf, fmt.Sprintf("shipment = \"%v\"", name))
	assert.Contains(t, tf, fmt.Sprintf("group    = \"%v\"", group))

	//shipment_env
	assert.Contains(t, tf, fmt.Sprintf("resource \"harbor_shipment_env\" \"%v\" {", env))
	assert.Contains(t, tf, fmt.Sprintf("environment = \"%v\"", env))
	assert.Contains(t, tf, fmt.Sprintf("barge       = \"%v\"", barge))
	assert.Contains(t, tf, fmt.Sprintf("replicas    = %v", replicas))

	//shouldn't contain "primary" since there's only 1 container
	assert.NotContains(t, tf, "primary")

	//shouldn't contain "log_shipping"
	assert.NotContains(t, tf, "log_shipping")

	//output
	assert.Contains(t, tf, fmt.Sprintf("value = \"${harbor_shipment_env.%v.dns_name}\"", env))
}

func TestGenerateTerraformSourceCodeLogShipping(t *testing.T) {

	shipmentJSON := getSampleShipmentJSONForLogShipping()

	//update json with test values
	name := "mss-poc-app"
	env := "dev"
	barge := "digital-sandbox"
	replicas := 2
	group := "mss"
	foo := "bar"
	project := "project"
	property := "property"
	product := "product"
	shipLogs := "logzio"
	logsEndpoint := "http://endpoint"
	region := "us-east-1"
	accessKey := "access"
	secretKey := "secret"
	domainName := "domain"
	queueName := "queue"
	containerLevel := "CONTAINER_LEVEL"
	container := "web"
	healthcheckTimeout := 10
	healthcheckInterval := 100

	shipmentJSON = strings.Replace(shipmentJSON, "${name}", name, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${env}", env, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${barge}", barge, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${replicas}", strconv.Itoa(replicas), 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${group}", group, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${foo}", foo, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${property}", property, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${product}", product, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${project}", project, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${shipLogs}", shipLogs, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${logsEndpoint}", logsEndpoint, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${domainName}", domainName, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${accessKey}", accessKey, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${secretKey}", secretKey, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${region}", region, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${queueName}", queueName, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${containerLevel}", containerLevel, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${container}", container, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${healthcheckTimeout}", strconv.Itoa(healthcheckTimeout), 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${healthcheckInterval}", strconv.Itoa(healthcheckInterval), 1)
	t.Log(shipmentJSON)

	//deserialize shipit json
	var shipmentEnv ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipmentEnv)
	if err != nil {
		t.Error(err)
	}

	//package the data to make it easy for rendering
	harborCompose := transformShipmentToHarborCompose(&shipmentEnv)
	data := getTerraformData(&shipmentEnv, &harborCompose)

	//generate code
	tf := generateTerraformSourceCode(data)
	t.Log(tf)

	//log_shipping
	assert.Contains(t, tf, "log_shipping")
	assert.Contains(t, tf, fmt.Sprintf("provider = \"%v\"", shipLogs))
	assert.Contains(t, tf, fmt.Sprintf("endpoint = \"%v\"", logsEndpoint))
	assert.Contains(t, tf, fmt.Sprintf("aws_elasticsearch_domain_name = \"%v\"", domainName))
	assert.Contains(t, tf, fmt.Sprintf("aws_access_key = \"%v\"", accessKey))
	assert.Contains(t, tf, fmt.Sprintf("aws_secret_key = \"%v\"", secretKey))
	assert.Contains(t, tf, fmt.Sprintf("aws_region = \"%v\"", region))
	assert.Contains(t, tf, fmt.Sprintf("sqs_queue_name = \"%v\"", queueName))
}
