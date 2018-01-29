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
After you have a main.tf, you can use terraform to import the existing state in order to start managing it with terraform.`,
	Example: `harbor-compose terraform
terraform import harbor_shipment.app my-app
terraform import harbor_shipment_env.dev my-app::dev
`,
	Run:    terraform,
	PreRun: preRunHook,
}

const (
	tfFile = "main.tf"
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
	generateAndWriteTerraformSource(shipmentObject, &harborCompose, true, false, "")

	fmt.Println("done")
}

func generateAndWriteTerraformSource(shipmentEnvironment *ShipmentEnvironment, harborCompose *HarborCompose, printUsage bool, outputRole bool, samlUser string) {

	//package the data to make it easy for rendering
	data := getTerraformData(shipmentEnvironment, harborCompose, outputRole, samlUser)

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

func getTerraformData(shipmentEnvironment *ShipmentEnvironment, harborCompose *HarborCompose, role bool, samlUser string) *terraformShipmentEnvironment {

	composeShipment := harborCompose.Shipments[shipmentEnvironment.ParentShipment.Name]
	monitored := true
	if composeShipment.EnableMonitoring != nil {
		monitored = *composeShipment.EnableMonitoring
	}

	awsProfile := "default"
	awsProfileEnv := os.Getenv("AWS_PROFILE")
	if awsProfileEnv != "" {
		awsProfile = awsProfileEnv
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
		Role:        role,
		AwsProfile:  awsProfile,
		SamlUser:    samlUser,
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
  shipment    = "${harbor_shipment.app.shipment}"
  environment = "{{ .Env }}"
	barge       = "{{ .Barge }}"
  replicas    = {{ .Replicas }}
	monitored   = {{ .Monitored }}
	{{ if .Role }}iam_role    = "${aws_iam_role.app_role.arn}"{{ end }}	
	{{ range .Containers }}
	container {
		{{ if .Primary }}name = "{{ .Name }}"{{ else }}
		name    = "{{ .Name }}"
		primary = false		
		{{ end }}
		{{ range .Ports }}
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
{{ if .Role }}
provider "aws" {
  profile = "{{ .AwsProfile }}"
}

# todo: fill out custom aws role policy
resource "aws_iam_role_policy" "aws_access" {
  name = "{{ .Shipment }}-{{ .Env }}"
  role = "${aws_iam_role.app_role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [],
      "Resource": []
    }
  ]
}
EOF
}

resource "aws_iam_role" "app_role" {
  name               = "{{ .Shipment }}-{{ .Env }}"
  assume_role_policy = "${data.aws_iam_policy_document.harbor_policy.json}"
}

module "harbor" {
  source = "git::ssh://git@github.com/turnercode/harbor-terraform?ref=v1.6"
  barge  = "{{ .Barge }}"
}

data "aws_caller_identity" "current" {}

# allow role to be assumed by harbor and saml user (for local dev)
data "aws_iam_policy_document" "harbor_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }

    principals {
      type = "AWS"

      identifiers = [
        "${module.harbor.iam_role}",
        "arn:aws:sts::${data.aws_caller_identity.current.account_id}:assumed-role/{{ .SamlUser }}",
      ]
    }
  }
}
{{ end }}
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
