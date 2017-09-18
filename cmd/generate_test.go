package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

//tests generating a docker-compose.yml from an existing shipment
func TestTransformShipmentToDockerCompose(t *testing.T) {

	//define a ShipmentEnvironment
	shipmentJSON := `
{
  "enableMonitoring": true,
  "name": "dev",
  "parentShipment": {
    "name": "mss-poc-app",
    "group": "mss",
    "envVars": [
      {
        "type": "basic",
        "value": "adds",
        "name": "CUSTOMER"
      },
      {
        "type": "basic",
        "value": "mss-poc-app",
        "name": "PRODUCT"
      },
      {
        "type": "basic",
        "value": "mss-poc-app",
        "name": "PROJECT"
      },
      {
        "type": "basic",
        "value": "mss",
        "name": "PROPERTY"
      }
    ]
  },
  "envVars": [
    {
      "type": "basic",
      "value": "bar",
      "name": "FOO"
    }
  ],
  "providers": [
    {
      "replicas": 2,
      "barge": "corp-sandbox",
      "name": "ec2",
      "envVars": []
    }
  ],
  "containers": [
    {
      "image": "${image}",
      "name": "${service}",
      "envVars": [
        {
          "type": "basic",
          "value": "${healthCheck}",
          "name": "HEALTHCHECK"
        }
      ],
      "ports": [
        {
          "protocol": "http",
          "healthcheck": "${healthCheck}",
          "external": true,
          "primary": true,
          "public_vip": false,
          "enable_proxy_protocol": false,
          "ssl_arn": "",
          "ssl_management_type": "iam",
          "healthcheck_timeout": 1,
          "public_port": ${publicPort},
          "value": ${containerPort},
          "name": "PORT"
        }
      ]
    }
  ]
}	
`

	//update json with test values
	service := "mss-poc-app"
	image := "quay.io/turner/mss-poc-app:1.0.0"
	publicPort := "80"
	containerPort := "3000"
	healthCheck := "/hc"
	foo := "bar"

	shipmentJSON = strings.Replace(shipmentJSON, "${service}", service, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${image}", image, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${publicPort}", publicPort, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${containerPort}", containerPort, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${healthCheck}", healthCheck, 1)

	//deserialize json
	var shipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipment)
	if err != nil {
		log.Fatal(err)
	}

	//test
	dockerCompose := transformShipmentToDockerCompose(&shipment, nil)

	//debug
	data, _ := yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//assertions
	assert.Equal(t, 1, len(dockerCompose.Services))
	composeService := dockerCompose.Services[service]
	assert.NotNil(t, composeService)
	assert.Equal(t, image, composeService.Image)
	assert.Equal(t, fmt.Sprintf("%v:%v", publicPort, containerPort), composeService.Ports[0])
	assert.Equal(t, containerPort, composeService.Environment["PORT"])
	assert.Equal(t, healthCheck, composeService.Environment["HEALTHCHECK"])
	assert.Equal(t, foo, composeService.Environment["FOO"])
}

//tests generating a docker-compose.yml from an existing shipment with multiple containers
func TestTransformShipmentToDockerComposeMultiContainer(t *testing.T) {

	//define a ShipmentEnvironment
	shipmentJSON := `
{
  "enableMonitoring": true,
  "name": "dev",
  "parentShipment": {
    "name": "mss-poc-app",
    "group": "mss",
    "envVars": [
      {
        "type": "basic",
        "value": "adds",
        "name": "CUSTOMER"
      },
      {
        "type": "basic",
        "value": "mss-poc-app",
        "name": "PRODUCT"
      },
      {
        "type": "basic",
        "value": "mss-poc-app",
        "name": "PROJECT"
      },
      {
        "type": "basic",
        "value": "mss",
        "name": "PROPERTY"
      }
    ]
  },
  "envVars": [
    {
      "type": "basic",
      "value": "bar",
      "name": "FOO"
    }
  ],
  "providers": [
    {
      "replicas": 2,
      "barge": "corp-sandbox",
      "name": "ec2",
      "envVars": []
    }
  ],
  "containers": [
    {
      "image": "${image}",
      "name": "${service}",
      "envVars": [
        {
          "type": "basic",
          "value": "${healthCheck}",
          "name": "HEALTHCHECK"
        }
      ],
      "ports": [
        {
          "protocol": "http",
          "healthcheck": "${healthCheck}",
          "external": true,
          "primary": true,
          "public_vip": false,
          "enable_proxy_protocol": false,
          "ssl_arn": "",
          "ssl_management_type": "iam",
          "public_port": ${publicPort},
          "value": ${containerPort},
          "name": "PORT"
        }
      ]
    },
    {
      "image": "${image}",
      "name": "${service2}",
      "envVars": [
        {
          "type": "basic",
          "value": "${healthCheck}",
          "name": "HEALTHCHECK"
        }
      ],
      "ports": [
        {
          "protocol": "http",
          "healthcheck": "${healthCheck}",
          "external": true,
          "primary": true,
          "public_vip": false,
          "enable_proxy_protocol": false,
          "ssl_arn": "",
          "ssl_management_type": "iam",
          "public_port": ${publicPort},
          "value": ${containerPort},
          "name": "PORT"
        }
      ]
    }    
  ]
}	
`
	//update json with test values
	service := "mss-poc-app"
	service2 := "mss-poc-app2"
	image := "quay.io/turner/mss-poc-app:1.0.0"
	publicPort := "80"
	containerPort := "3000"
	healthCheck := "/hc"
	foo := "bar"

	shipmentJSON = strings.Replace(shipmentJSON, "${service}", service, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${service2}", service2, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${image}", image, 2)
	shipmentJSON = strings.Replace(shipmentJSON, "${publicPort}", publicPort, 2)
	shipmentJSON = strings.Replace(shipmentJSON, "${containerPort}", containerPort, 2)
	shipmentJSON = strings.Replace(shipmentJSON, "${healthCheck}", healthCheck, 4)

	//deserialize json
	var shipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipment)
	if err != nil {
		log.Fatal(err)
	}

	//test
	dockerCompose := transformShipmentToDockerCompose(&shipment, nil)

	//debug
	data, _ := yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//assertions
	assert.Equal(t, 2, len(dockerCompose.Services))

	composeService := dockerCompose.Services[service]
	assert.NotNil(t, composeService)
	assert.Equal(t, image, composeService.Image)
	assert.Equal(t, fmt.Sprintf("%v:%v", publicPort, containerPort), composeService.Ports[0])
	assert.Equal(t, containerPort, composeService.Environment["PORT"])
	assert.Equal(t, healthCheck, composeService.Environment["HEALTHCHECK"])
	assert.Equal(t, foo, composeService.Environment["FOO"])

	composeService2 := dockerCompose.Services[service2]
	assert.NotNil(t, composeService2)
	assert.Equal(t, image, composeService2.Image)
	assert.Equal(t, fmt.Sprintf("%v:%v", publicPort, containerPort), composeService2.Ports[0])
	assert.Equal(t, containerPort, composeService2.Environment["PORT"])
	assert.Equal(t, healthCheck, composeService2.Environment["HEALTHCHECK"])
	assert.Equal(t, foo, composeService2.Environment["FOO"])
}

//tests generating a harbor-compose.yml from an existing shipment
func TestTransformShipmentToHarborCompose(t *testing.T) {
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
	t.Log(shipmentJSON)

	//deserialize shipit json
	var shipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipment)
	if err != nil {
		log.Fatal(err)
	}

	//convert shipit model to harbor-compose
	harborCompose, hiddenEnvVars := transformShipmentToHarborCompose(&shipment)

	//convert shipit model to docker-compose
	dockerCompose := transformShipmentToDockerCompose(&shipment, hiddenEnvVars)

	//debug
	data, _ := yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//debug
	data, _ = yaml.Marshal(harborCompose)
	t.Log(string(data))

	//assertions
	assert.Equal(t, 1, len(harborCompose.Shipments))
	composeShipment := harborCompose.Shipments[name]
	assert.NotNil(t, composeShipment)
	assert.Equal(t, group, composeShipment.Group)
	assert.Equal(t, barge, composeShipment.Barge)
	assert.Equal(t, env, composeShipment.Env)
	assert.Equal(t, replicas, composeShipment.Replicas)
	assert.Equal(t, project, composeShipment.Project)
	assert.Equal(t, property, composeShipment.Property)
	assert.Equal(t, product, composeShipment.Product)
	assert.Equal(t, 1, len(composeShipment.Containers))
	assert.Equal(t, true, *composeShipment.EnableMonitoring)

	//IgnoreImageVersion should default to false
	assert.Equal(t, false, composeShipment.IgnoreImageVersion)

	//both container-level and env-level shipit envvars should get added to docker-compose and not harbor-compose
	assert.Equal(t, containerLevel, dockerCompose.Services["web"].Environment[containerLevel])
	assert.Equal(t, envLevel, dockerCompose.Services["web"].Environment[envLevel])
	assert.NotEqual(t, envLevel, composeShipment.Environment[envLevel])
}

func getSampleShipmentJSON() string {
	return `
{
  "name": "${env}",
  "enableMonitoring": true,
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
      "value": "${envLevel}",
      "name": "ENV_LEVEL"
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

func getSampleShipmentJSONForValidation() string {
	return `
{
  "name": "${env}",
  "parentShipment": {
    "name": "mss-poc-app",
    "group": "group",
    "envVars": [
      {
        "type": "basic",
        "value": "customer",
        "name": "CUSTOMER"
      },
      {
        "type": "basic",
        "value": "product",
        "name": "PRODUCT"
      },
      {
        "type": "basic",
        "value": "project",
        "name": "PROJECT"
      },
      {
        "type": "basic",
        "value": "property",
        "name": "PROPERTY"
      }
    ]
  },
  "envVars": [
    {
      "type": "basic",
      "value": "envLevel",
      "name": "ENV_LEVEL"
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
          "value": "containerLevel",
          "name": "CONTAINER_LEVEL"
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

func getSampleShipmentRestartJSON() string {
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
      "value": "${envLevel}",
      "name": "ENV_LEVEL"
		},
    {
      "type": "basic",
      "value": "${envLevel}",
      "name": "HC_RESTART"
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

//tests filtering of HC_RESTART
func TestTransformShipmentToHarborComposeRestartEnv(t *testing.T) {
	shipmentJSON := getSampleShipmentRestartJSON()

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
	t.Log(shipmentJSON)

	//deserialize shipit json
	var shipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipment)
	if err != nil {
		log.Fatal(err)
	}

	//convert shipit model to harbor-compose
	harborCompose, hiddenEnvVars := transformShipmentToHarborCompose(&shipment)

	//convert shipit model to docker-compose
	dockerCompose := transformShipmentToDockerCompose(&shipment, hiddenEnvVars)

	//debug
	data, _ := yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//debug
	data, _ = yaml.Marshal(harborCompose)
	t.Log(string(data))

	//assertions
	assert.Equal(t, 1, len(harborCompose.Shipments))
	composeShipment := harborCompose.Shipments[name]
	assert.NotNil(t, composeShipment)
	assert.Equal(t, group, composeShipment.Group)
	assert.Equal(t, barge, composeShipment.Barge)
	assert.Equal(t, env, composeShipment.Env)
	assert.Equal(t, replicas, composeShipment.Replicas)
	assert.Equal(t, project, composeShipment.Project)
	assert.Equal(t, property, composeShipment.Property)
	assert.Equal(t, product, composeShipment.Product)
	assert.Equal(t, 1, len(composeShipment.Containers))

	//IgnoreImageVersion should default to false
	assert.Equal(t, false, composeShipment.IgnoreImageVersion)

	//both container-level and env-level shipit envvars should get added to docker-compose and not harbor-compose
	assert.Equal(t, containerLevel, dockerCompose.Services[container].Environment[containerLevel])
	assert.Equal(t, envLevel, dockerCompose.Services[container].Environment[envLevel])
	assert.NotEqual(t, envLevel, composeShipment.Environment[envLevel])

	//HC_RESTART should not get added to yaml files
	hcRestart := "HC_RESTART"
	assert.Equal(t, "", dockerCompose.Services[container].Environment[hcRestart], "expecting %v to not be added", hcRestart)
	assert.Equal(t, "", composeShipment.Environment[hcRestart], "expecting %v to not be added", hcRestart)
}

//tests generating a docker-compose.yml from an existing shipment with hidden env vars
func TestTransformShipmentToDockerComposeWithHiddenEnvVar(t *testing.T) {

	//define a ShipmentEnvironment
	shipmentJSON := `
  {
    "enableMonitoring": true,
    "name": "dev",
    "parentShipment": {
      "name": "mss-poc-app",
      "group": "mss",
      "envVars": [
        {
          "type": "basic",
          "value": "adds",
          "name": "CUSTOMER"
        },
        {
          "type": "basic",
          "value": "mss-poc-app",
          "name": "PRODUCT"
        },
        {
          "type": "basic",
          "value": "mss-poc-app",
          "name": "PROJECT"
        },
        {
          "type": "basic",
          "value": "mss",
          "name": "PROPERTY"
        }
      ]
    },
    "envVars": [
      {
        "type": "basic",
        "value": "bar",
        "name": "FOO"
      }
    ],
    "providers": [
      {
        "replicas": 2,
        "barge": "corp-sandbox",
        "name": "ec2",
        "envVars": []
      }
    ],
    "containers": [
      {
        "image": "${image}",
        "name": "${service}",
        "envVars": [
          {
            "type": "basic",
            "value": "${healthCheck}",
            "name": "HEALTHCHECK"
          },
          {
            "type": "hidden",
            "value": "${hiddenValue}",
            "name": "HIDDEN"
          }          
        ],
        "ports": [
          {
            "protocol": "http",
            "healthcheck": "${healthCheck}",
            "external": true,
            "primary": true,
            "public_vip": false,
            "enable_proxy_protocol": false,
            "ssl_arn": "",
            "ssl_management_type": "iam",
            "healthcheck_timeout": 1,
            "public_port": ${publicPort},
            "value": ${containerPort},
            "name": "PORT"
          }
        ]
      }
    ]
  }	
  `

	//update json with test values
	service := "mss-poc-app"
	image := "quay.io/turner/mss-poc-app:1.0.0"
	publicPort := "80"
	containerPort := "3000"
	healthCheck := "/hc"
	hiddenValue := "some-hidden-value"

	shipmentJSON = strings.Replace(shipmentJSON, "${service}", service, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${image}", image, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${publicPort}", publicPort, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${containerPort}", containerPort, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${healthCheck}", healthCheck, 1)
	shipmentJSON = strings.Replace(shipmentJSON, "${hiddenValue}", hiddenValue, 1)

	//deserialize json
	var shipment ShipmentEnvironment
	err := json.Unmarshal([]byte(shipmentJSON), &shipment)
	if err != nil {
		log.Fatal(err)
	}

	//test
	hiddenEnvVars := map[string]string{}
	dockerCompose := transformShipmentToDockerCompose(&shipment, hiddenEnvVars)

	//debug
	data, _ := yaml.Marshal(dockerCompose)
	t.Log(string(data))

	//assertions
	composeService := dockerCompose.Services[service]

	//check hidden.env
	if len(composeService.EnvFile) != 1 {
		assert.FailNow(t, "env_file should contain 1 value")
	}
	envFileName := composeService.EnvFile[0]
	assert.Equal(t, "hidden.env", envFileName, "env_file should contain hidden.env")
}
