package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/spf13/cobra"
)

// terraformCmd represents the terraform command
var terraformCmd = &cobra.Command{
	Use:   "terraform",
	Short: "generates terraform source code from an existing shipment",
	Long: `generates terraform source code from an existing shipment

The terraform command outputs terraform files (main.tf) from an existing shipment environment.
After you have a main.tf, you can use terraform to import the existing state.

Example:
harbor-compose terraform my-shipment dev
terraform import harbor_shipment.app my-app
terraform import harbor_shipment_env.dev my-app::dev
`,
	Run:    terraform,
	PreRun: preRunHook,
}

const (
	tfFile = "main.tf"
)

//log shipping env vars
const (
	envVarNameShipLogs     = "SHIP_LOGS"
	envVarNameLogsEndpoint = "LOGS_ENDPOINT"
	envVarNameAccessKey    = "LOGS_ACCESS_KEY"
	envVarNameSecretKey    = "LOGS_SECRET_KEY"
	envVarNameDomainName   = "LOGS_DOMAIN_NAME"
	envVarNameRegion       = "LOGS_REGION"
	envVarNameQueueName    = "LOGS_QUEUE_NAME"
)

func init() {
	RootCmd.AddCommand(terraformCmd)
}

func terraform(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		log.Fatal("shipment and environment arguments are required. ex: harbor-compose terraform my-shipment dev")
	}

	username, token, err := Login()
	check(err)

	shipment := args[0]
	env := args[1]

	if Verbose {
		log.Printf("fetching shipment...")
	}
	shipmentObject := GetShipmentEnvironment(username, token, shipment, env)
	if shipmentObject == nil {
		fmt.Println(messageShipmentEnvironmentNotFound)
		return
	}

	//convert a Shipment object into a HarborCompose object
	harborCompose := transformShipmentToHarborCompose(shipmentObject)

	//generate a main.tf and write it to disk
	generateAndWriteTerraformSource(shipmentObject, &harborCompose, true)

	fmt.Println("done")
}

func generateAndWriteTerraformSource(shipmentEnvironment *ShipmentEnvironment, harborCompose *HarborCompose, printUsage bool) {

	//package the data to make it easy for rendering
	data := getTerraformData(shipmentEnvironment, harborCompose)

	//translate a shipit shipment environment into terraform source
	tfCode := generateTerraformSourceCode(data)

	//prompt if the file already exist
	yes := true
	if _, err := os.Stat(tfFile); err == nil { //exists
		fmt.Printf(tfFile + " already exists. Overwrite? ")
		yes = askForConfirmation()
	}
	if yes {
		err := ioutil.WriteFile(tfFile, []byte(tfCode), 0644)
		check(err)
		fmt.Println("wrote " + tfFile)
		if printUsage {
			fmt.Println()
			fmt.Println("to start using terraform, run the following commands to import current state:")
			fmt.Println()
			fmt.Println("terraform init")
			fmt.Printf("terraform import harbor_shipment.app %v\n", shipmentEnvironment.ParentShipment.Name)
			fmt.Printf("terraform import harbor_shipment_env.%v %v::%v\n", shipmentEnvironment.Name, shipmentEnvironment.ParentShipment.Name, shipmentEnvironment.Name)
			fmt.Println()
		}
	}
}

func getTerraformData(shipmentEnvironment *ShipmentEnvironment, harborCompose *HarborCompose) *terraformShipmentEnvironment {

	composeShipment := harborCompose.Shipments[shipmentEnvironment.ParentShipment.Name]
	monitored := true
	if composeShipment.EnableMonitoring != nil {
		monitored = *composeShipment.EnableMonitoring
	}

	result := terraformShipmentEnvironment{
		Shipment:    shipmentEnvironment.ParentShipment.Name,
		Env:         shipmentEnvironment.Name,
		Group:       composeShipment.Group,
		Barge:       composeShipment.Barge,
		Replicas:    composeShipment.Replicas,
		Monitored:   monitored,
		Containers:  []terraformContainer{},
		LogShipping: terraformLogShipping{},
	}

	for _, c := range shipmentEnvironment.Containers {
		container := terraformContainer{
			Name:  c.Name,
			Ports: []terraformPort{},
		}

		for _, p := range c.Ports {
			port := terraformPort{
				Healthcheck:         p.Healthcheck,
				HealthcheckInterval: 10,
				HealthcheckTimeout:  1,
				Value:               p.Value,
				Protocol:            p.Protocol,
				External:            p.External,
				PublicVip:           p.PublicVip,
				PublicPort:          p.PublicPort,
				EnableProxyProtocol: p.EnableProxyProtocol,
				SslArn:              p.SslArn,
				SslManagementType:   p.SslManagementType,
			}

			if p.HealthcheckInterval != nil {
				port.HealthcheckInterval = *p.HealthcheckInterval
			}
			if p.HealthcheckTimeout != nil {
				port.HealthcheckTimeout = *p.HealthcheckTimeout
			}

			//set container as primary since it contains the shipment/env's primary port
			//and there can only be 1 per shipment/env
			if p.Primary {
				container.Primary = true
			}

			container.Ports = append(container.Ports, port)
		}
		result.Containers = append(result.Containers, container)
	}

	//look for log shipping env vars
	if envvar := findEnvVar(envVarNameShipLogs, shipmentEnvironment.EnvVars); envvar != (EnvVarPayload{}) {
		result.LogShipping.IsSpecified = true
		result.LogShipping.Provider = envvar.Value
	}

	if envvar := findEnvVar(envVarNameLogsEndpoint, shipmentEnvironment.EnvVars); envvar != (EnvVarPayload{}) {
		result.LogShipping.Endpoint = envvar.Value
	}

	if envvar := findEnvVar(envVarNameDomainName, shipmentEnvironment.EnvVars); envvar != (EnvVarPayload{}) {
		result.LogShipping.AwsElasticsearchDomainName = envvar.Value
	}

	if envvar := findEnvVar(envVarNameRegion, shipmentEnvironment.EnvVars); envvar != (EnvVarPayload{}) {
		result.LogShipping.AwsRegion = envvar.Value
	}

	if envvar := findEnvVar(envVarNameAccessKey, shipmentEnvironment.EnvVars); envvar != (EnvVarPayload{}) {
		result.LogShipping.AwsAccessKey = envvar.Value
	}

	if envvar := findEnvVar(envVarNameSecretKey, shipmentEnvironment.EnvVars); envvar != (EnvVarPayload{}) {
		result.LogShipping.AwsSecretKey = envvar.Value
	}

	if envvar := findEnvVar(envVarNameQueueName, shipmentEnvironment.EnvVars); envvar != (EnvVarPayload{}) {
		result.LogShipping.SqsQueueName = envvar.Value
	}

	return &result
}

func generateTerraformSourceCode(data *terraformShipmentEnvironment) string {

	tf := `# generated by harbor-compose

provider "harbor" {
	credentials = "${file("~/.harbor/credentials")}"
}

resource "harbor_shipment" "app" {
  shipment = "{{ .Shipment }}"
  group    = "{{ .Group }}"
}

resource "harbor_shipment_env" "{{ .Env }}" {
  shipment    = "${harbor_shipment.app.id}"
  environment = "{{ .Env }}"
  barge       = "{{ .Barge }}"
  replicas    = {{ .Replicas }}
	monitored   = {{ .Monitored }}
	{{ range .Containers }}
	container {
		{{ if .Primary }}
		name = "{{ .Name }}"
		{{ else }}
		name    = "{{ .Name }}"
		primary = false		
		{{ end }}{{ range .Ports }}
    port {
			protocol              = "{{ .Protocol }}"
			public_port           = {{ .PublicPort }}
			value                 = {{ .Value }}
			healthcheck           = "{{ .Healthcheck }}"
			healthcheck_timeout   = {{ .HealthcheckTimeout }}
			healthcheck_interval  = {{ .HealthcheckInterval }}
			external              = {{ .External }}
			public                = {{ .PublicVip }}
			enable_proxy_protocol = {{ .EnableProxyProtocol }}
			ssl_management_type   = "{{ .SslManagementType }}"
			ssl_arn               = "{{ .SslArn }}"
		}{{ end }}
	}{{ end }}
	{{ if .LogShipping.IsSpecified }}
  log_shipping {
    provider = "{{ .LogShipping.Provider }}"
    endpoint = "{{ .LogShipping.Endpoint }}"
    aws_elasticsearch_domain_name = "{{ .LogShipping.AwsElasticsearchDomainName }}"
    aws_region = "{{ .LogShipping.AwsRegion }}"
    aws_access_key = "{{ .LogShipping.AwsAccessKey }}"
    aws_secret_key = "{{ .LogShipping.AwsSecretKey }}"
    sqs_queue_name = "{{ .LogShipping.SqsQueueName }}"
	}{{ end }}	
}

output "dns_name" {
  value = "${harbor_shipment_env.{{ .Env }}.dns_name}"
}

output "lb_name" {
  value = "${harbor_shipment_env.{{ .Env }}.lb_name}"
}

output "lb_dns_name" {
  value = "${harbor_shipment_env.{{ .Env }}.lb_dns_name}"
}
`

	//create a formatted template
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 0, 2, ' ', tabwriter.DiscardEmptyColumns)
	tmpl, err := template.New("tf").Parse(tf)
	check(err)
	fmt.Fprintln(w)

	//execute the template with the data
	err = tmpl.Execute(w, data)
	check(err)
	w.Flush()

	return buf.String()
}
