package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	getter "github.com/hashicorp/go-getter"
)

//writes customized migration template and returns the directory
func migrateToEcsFargate(shipmentEnv *ShipmentEnvironment, harborCompose *HarborCompose) (string, *ecsTerraformShipmentEnvironment) {

	//fetch terraform template
	repoDir := downloadTerraformTemplate()
	debug("downloaded to: " + repoDir)

	baseDir, envDir, baseDirInstalled := installTerraformTemplate(repoDir, shipmentEnv.Name)
	debug("environment installed to: " + envDir)

	//translate harbor data into aws ecs fargate data
	data := translateShipmentEnvironmentToEcsTerraform(shipmentEnv, harborCompose)

	//generate terraform.tfvars for base (if installed)
	if baseDirInstalled {
		baseTfVars := getTfVarsForBase(data)
		debug(baseTfVars)
		baseTfVarsFile := filepath.Join(baseDir, "terraform.tfvars")
		debug("writing " + baseTfVarsFile)
		err := ioutil.WriteFile(baseTfVarsFile, []byte(baseTfVars), 0644)
		check(err)
	}

	//generate terraform.tfvars for env
	envTfVars := getTfVarsForEnv(data)
	debug(envTfVars)
	envTfVarsFile := filepath.Join(envDir, "terraform.tfvars")
	debug("writing " + envTfVarsFile)
	err := ioutil.WriteFile(envTfVarsFile, []byte(envTfVars), 0644)
	check(err)

	//update tf backend in main.tf
	mainTfFile := filepath.Join(envDir, "main.tf")
	fileBits, err := ioutil.ReadFile(mainTfFile)
	check(err)
	maintf := updateTerraformBackend(string(fileBits), data)
	err = ioutil.WriteFile(mainTfFile, []byte(maintf), 0644)
	check(err)

	//configure ALB listeners and security groups based on harbor configuration:
	//  delete lb-https.tf for http only
	//  delete lb-http.tf for https only
	//  do nothing for http and https
	if data.HTTPPort != nil && data.HTTPSPort == nil {
		debug("deleting lb-https.tf")
		err = os.Remove(filepath.Join(envDir, "lb-https.tf"))
		check(err)
	}
	if data.HTTPPort == nil && data.HTTPSPort != nil {
		debug("deleting lb-http.tf")
		err = os.Remove(filepath.Join(envDir, "lb-http.tf"))
		check(err)

		//update listener dependency in ecs.tf
		tfFile := filepath.Join(envDir, "ecs.tf")
		fileBits, err := ioutil.ReadFile(tfFile)
		check(err)
		fileContents := updateEcsListenerDep(string(fileBits))
		err = ioutil.WriteFile(tfFile, []byte(fileContents), 0644)
		check(err)
	}

	//update lb-https.tf with support for iam server certificates
	if data.HTTPSPort != nil && data.HTTPSPort.SslManagementType == "iam" {

		//lookup cert name prefix from harbor
		certNamePrefix := getSslCertNamePrefix(data.Shipment, data.Env)

		//re-write lb-https.tf
		tfFile := filepath.Join(envDir, "lb-https.tf")
		fileBits, err := ioutil.ReadFile(tfFile)
		check(err)
		fileContents := updateHTTPSForIam(string(fileBits), certNamePrefix)
		err = ioutil.WriteFile(tfFile, []byte(fileContents), 0644)
		check(err)
	}

	//write doc-monitoring files
	downloadDocMonitoringFiles(envDir)

	//write migrate-image files
	downloadImageMigrationFiles(envDir)

	//configure log shipping for logz.io
	if data.LogShipping.Provider != "logzio" {
		debug("deleting logs-logzio files")
		err = os.Remove(filepath.Join(envDir, "logs-logzio.tf"))
		check(err)
		err = os.Remove(filepath.Join(envDir, "logs-logzio.zip"))
		check(err)
	}

	//delete autoscale-time for "prod*" environments
	if strings.Contains(data.Env, "prod") {
		debug("deleting autoscale-time files for prod environment: " + data.Env)
		err = os.Remove(filepath.Join(envDir, "autoscale-time.tf"))
		err = os.Remove(filepath.Join(envDir, "autoscale-time.zip"))
	}

	//write a fargate.yml for the cli
	fargateYml := getFargateYaml(data.App, data.Env)
	fargateYmlFile := filepath.Join(envDir, "fargate.yml")
	debug("writing " + fargateYmlFile)
	err = ioutil.WriteFile(fargateYmlFile, []byte(fargateYml), 0644)
	check(err)

	return envDir, data
}

func getSslCertNamePrefix(shipment string, env string) string {
	lb, err := getLoadBalancerStatus(shipment, env)
	check(err)
	return lb.Name
}

func updateHTTPSForIam(tf string, certNamePrefix string) string {
	//replace `certificate_arn   = "${var.certificate_arn}"`
	//with `certificate_arn = "${data.aws_iam_server_certificate.app.arn}"`
	tmp := strings.Split(tf, "\n")
	newTf := ""
	for _, line := range tmp {
		updatedLine := line
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, `certificate_arn`) {
			updatedLine = `  certificate_arn   = "${data.aws_iam_server_certificate.app.arn}"`
		}
		newTf += updatedLine + "\n"
	}

	//add aws_iam_server_certificate data source
	newTf += fmt.Sprintf(`
data "aws_iam_server_certificate" "app" {
  name_prefix = "%s"
  latest      = true
}
	`, certNamePrefix)

	//delete variable "certificate_arn"
	newTf = strings.Replace(newTf, `# The ARN for the SSL certificate`, "", -1)
	newTf = strings.Replace(newTf, `variable "certificate_arn" {}`, "", -1)

	return newTf
}

func getFargateYaml(shipment string, env string) string {
	return fmt.Sprintf(`cluster: %s-%s
service: %s-%s
`, shipment, env, shipment, env)
}

func updateTerraformBackend(maintf string, data *ecsTerraformShipmentEnvironment) string {
	//replace profile and bucket
	tmp := strings.Split(maintf, "\n")
	newMaintf := ""
	for _, line := range tmp {
		updatedLine := line
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, `profile = ""`) {
			updatedLine = fmt.Sprintf(`    profile = "%s"`, data.AwsProfile)
		}
		if strings.HasPrefix(trimmed, "bucket") {
			updatedLine = fmt.Sprintf(`    bucket  = "tf-state-%s"`, data.App)
		}
		if strings.HasPrefix(trimmed, "key") {
			updatedLine = fmt.Sprintf(`    key     = "%s.terraform.tfstate"`, data.Env)
		}

		newMaintf += updatedLine + "\n"
	}
	return newMaintf
}

func updateEcsListenerDep(maintf string) string {
	//replace "aws_alb_listener.http" with "aws_alb_listener.https"
	tmp := strings.Split(maintf, "\n")
	newMaintf := ""
	for _, line := range tmp {
		updatedLine := line
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, `"aws_alb_listener.http",`) {
			updatedLine = `    "aws_alb_listener.https"`
		}
		newMaintf += updatedLine + "\n"
	}
	return newMaintf
}

//fetches and installs the tf template and returns the output directory
func downloadTerraformTemplate() string {

	org := "turnerlabs"
	repo := "terraform-ecs-fargate"
	url := fmt.Sprintf("https://github.com/%s/%s/archive/%s.zip", org, repo, migrateTemplateTag)
	client := getter.Client{
		Src:  url,
		Dst:  "./",
		Mode: getter.ClientModeDir,
	}

	fmt.Println("downloading terraform template", url)
	err := client.Get()
	check(err)
	debug("done")

	repoDir, err := getter.SubdirGlob("./", repo+"*")
	check(err)

	return repoDir
}

//installs a template for the specified environment,
//indicating whether or not the base directory was installed
func installTerraformTemplate(repoDir string, environment string) (string, string, bool) {

	//create infrastructure directory (if not already there)
	infraDir := "infrastructure"
	fmt.Println("installing terraform template")
	if _, err := os.Stat(infraDir); os.IsNotExist(err) {
		debug("creating directory: " + infraDir)
		err = os.Mkdir(infraDir, 0755)
		check(err)
	} else {
		debug(infraDir + " already exists")
	}

	//copy over infrastructure/base (if not already there)
	baseDir := "base"
	sourceBaseDir := filepath.Join(repoDir, baseDir)
	destBaseDir := filepath.Join(infraDir, baseDir)
	createdBaseDir := false
	if _, err := os.Stat(destBaseDir); os.IsNotExist(err) {
		debug(fmt.Sprintf("copying %s to %s", sourceBaseDir, destBaseDir))
		err = copyDir(sourceBaseDir, destBaseDir)
		check(err)
		createdBaseDir = true
	} else {
		fmt.Println(destBaseDir + " already exists, ignoring")
	}

	//if environment directory exists, prompt to override, if no, then exit
	sourceEnvDir := filepath.Join(repoDir, "env", "dev")
	destEnvDir := filepath.Join(infraDir, "env", environment)
	if _, err := os.Stat(destEnvDir); err == nil {
		//exists
		fmt.Print(destEnvDir + " already exists. Overwrite? ")
		if askForConfirmation() {
			debug("deleting " + destEnvDir)
			//delete environment directory (all files)
			err = os.RemoveAll(destEnvDir)
			check(err)
		} else {
			os.Exit(-1)
		}
	} else {
		//doesn't exist
		debug(destEnvDir + " doesn't exist")
	}

	//env directory either doesn't exist or user wants to overwrite
	//copy repo/env/${env} -> ./infrastructure/env/${env}
	debug(fmt.Sprintf("copying %s to %s", sourceEnvDir, destEnvDir))
	err := copyDir(sourceEnvDir, destEnvDir)
	check(err)

	//finally, delete repo dir
	debug("deleting: " + repoDir)
	err = os.RemoveAll(repoDir)
	check(err)

	return destBaseDir, destEnvDir, createdBaseDir
}

func getBargeData(barge string) *Barge {
	bargeResults := GetBarges()
	for _, v := range bargeResults.Barges {
		if v.Name == barge {
			return &v
		}
	}
	return nil
}

func getContactEmailFromGroup(group string) string {
	result := ""
	groupData := GetGroup(group)

	//take the first admin
	if len(groupData.Admins) > 0 {
		result = groupData.Admins[0]
	} else if len(groupData.Users) > 0 {
		result = groupData.Users[0]
	} else if Verbose {
		fmt.Println("could not find user for group: " + group)
	}

	return result
}

func downloadFileFromBargeEndpoint(dir string, file string) {
	url := bargesURI("/{file}", param("file", file))
	client := getter.Client{
		Src:  url,
		Dst:  filepath.Join("./", dir, file),
		Mode: getter.ClientModeFile,
	}
	debug("downloading", url)
	err := client.Get()
	check(err)
}

func downloadDocMonitoringFiles(envDir string) {
	downloadFileFromBargeEndpoint(envDir, "doc-monitoring.tf")
	downloadFileFromBargeEndpoint(envDir, "doc-monitoring.tpl")
}

func downloadImageMigrationFiles(envDir string) {
	downloadFileFromBargeEndpoint(envDir, "migrate-image.tf")
	downloadFileFromBargeEndpoint(envDir, "migrate-image.tpl")
}

func translateShipmentEnvironmentToEcsTerraform(shipmentEnvironment *ShipmentEnvironment, harborCompose *HarborCompose) *ecsTerraformShipmentEnvironment {

	composeShipment := harborCompose.Shipments[shipmentEnvironment.ParentShipment.Name]
	monitored := true
	if composeShipment.EnableMonitoring != nil {
		monitored = *composeShipment.EnableMonitoring
	}

	generationTime := time.Now().Format("Jan 2, 2006 at 3:04pm (MST)")

	result := ecsTerraformShipmentEnvironment{
		Shipment:      shipmentEnvironment.ParentShipment.Name,
		App:           shipmentEnvironment.ParentShipment.Name,
		Env:           shipmentEnvironment.Name,
		Group:         composeShipment.Group,
		Replicas:      composeShipment.Replicas,
		Monitored:     monitored,
		LogShipping:   terraformLogShipping{},
		IamRole:       shipmentEnvironment.IamRole,
		Product:       composeShipment.Product,
		Project:       composeShipment.Project,
		Property:      composeShipment.Property,
		ContainerName: shipmentEnvironment.Containers[0].Name,
		OldImage:      shipmentEnvironment.Containers[0].Image,
		GeneratedDate: generationTime,
	}

	//override "Shipment" with --app, if specified
	if migrateAppName != "" {
		result.App = migrateAppName
	}

	//call groups api to get contact-email address
	result.ContactEmail = getContactEmailFromGroup(composeShipment.Group)
	result.AwsRole = migrateRole
	result.IamRoleIsSpecified = (result.IamRole != "")

	//set aws info based on barge or cli args
	if migrateAccountName != "" {
		result.AwsAccountName = migrateAccountName
		result.AwsAccountID = migrateAccountID
		result.AwsVpc = migrateVPC
		result.AwsPrivateSubnets = migratePrivateSubnets
		result.AwsPublicSubnets = migratePublicSubnets
		result.AwsRegion = "us-east-1"
	} else {

		//call barge api to get aws account/networking info
		barge := getBargeData(composeShipment.Barge)
		if barge == nil {
			check(fmt.Errorf("barge %s not found", composeShipment.Barge))
		}

		result.AwsAccountID = barge.AccountID
		result.AwsAccountName = barge.AccountName
		result.AwsVpc = barge.Vpc
		result.AwsPrivateSubnets = strings.Join(barge.PrivateSubnets, ",")
		result.AwsPublicSubnets = strings.Join(barge.PublicSubnets, ",")
		result.AwsRegion = "us-east-1"
	}

	//set result.AwsSamlRole and result.AwsProfile
	samlRole, profile := getAwsSamlRoleAndProfile(result.AwsAccountName, result.AwsRole)
	result.AwsSamlRole = samlRole
	result.AwsProfile = profile

	//convert current image to ecr image
	result.NewImage = migrateImage(shipmentEnvironment.Containers[0].Image, result.AwsAccountID, result.AwsRegion, result.App)

	//use first container
	container := shipmentEnvironment.Containers[0]

	for _, p := range container.Ports {
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

		//determine if the lb should be public if any ports have public_vip = true
		result.PublicLB = p.PublicVip

		//store reference to primary port
		if p.Primary {
			result.PrimaryPort = &port
			if p.Healthcheck == "" {
				check(fmt.Errorf("primary port missing health check"))
			}

			//adjust harbor defaults to aws defaults
			//https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_CreateTargetGroup.html

			if port.HealthcheckInterval < 5 {
				port.HealthcheckInterval = 5
			}
			if port.HealthcheckTimeout == 1 {
				port.HealthcheckTimeout = 5
			}

			//public port is optional and should default to the value
			if port.PublicPort == 0 {
				port.PublicPort = port.Value
			}
		}

		//store references to http and https ports
		//and update public port to value if not specified
		if port.Protocol == "http" || port.Protocol == "tcp" {
			result.HTTPPort = &port
			if port.PublicPort == 0 {
				port.PublicPort = port.Value
			}
		} else if p.Protocol == "https" {
			result.HTTPSPort = &port
			if port.PublicPort == 0 {
				port.PublicPort = port.Value
			}
		}
	}

	if Verbose {
		if result.HTTPPort != nil {
			log.Printf("http port = %v:%v \n", result.HTTPPort.PublicPort, result.HTTPPort.Value)
		} else {
			log.Println("http port is null")
		}
		if result.HTTPSPort != nil {
			log.Printf("https port = %v:%v \n", result.HTTPSPort.PublicPort, result.HTTPSPort.Value)
		} else {
			log.Println("https port is null")
		}
		if result.PrimaryPort != nil {
			log.Printf("primary port = %v, %v \n", result.PrimaryPort.Protocol, result.PrimaryPort.Healthcheck)
		} else {
			log.Println("primary port is null")
		}
	}

	//look for log shipping env vars
	if envvar := findEnvVar(envVarNameShipLogs, shipmentEnvironment.EnvVars); envvar != (EnvVarPayload{}) {
		result.LogShipping.IsSpecified = true
		result.LogShipping.Provider = envvar.Value
		debug("found log shipping provider: ", envvar.Value)
	}

	if envvar := findEnvVar(envVarNameLogsEndpoint, shipmentEnvironment.EnvVars); envvar != (EnvVarPayload{}) {
		result.LogShipping.Endpoint = envvar.Value

		//parse out logz token
		//ex: https://listener.logz.io:8071?token=xyz
		tmp := strings.Split(envvar.Value, "=")
		result.LogzToken = tmp[1]
		debug(result.LogzToken)
	}

	return &result
}

func getAwsSamlRoleAndProfile(accountName string, role string) (string, string) {
	//set saml_role based on convention
	samlRole := fmt.Sprintf("%s-%s", accountName, role)

	//accomodate legacy accounts
	//prefix saml role with "aws-", if missing
	if !strings.HasPrefix(samlRole, "aws-") {
		samlRole = "aws-" + samlRole
	}

	//use default profile name if profile arg not specified
	profile := ""
	if migrateProfile != "" {
		profile = migrateProfile
	} else {
		profile = fmt.Sprintf("%s:%s", accountName, samlRole)
	}

	return samlRole, profile
}

func getTfVarsForBase(data *ecsTerraformShipmentEnvironment) string {

	tf := `# generated by harbor-compose on {{ .GeneratedDate }}

region = "{{ .AwsRegion }}"

aws_profile = "{{ .AwsProfile }}"
	
saml_role = "{{ .AwsSamlRole }}"
	
tags = {
	application      = "{{ .App }}"
	environment      = "prod"
	team             = "{{ .Group }}"
	customer         = "{{ .Property }}"
	contact-email    = "{{ .ContactEmail }}"
	product          = "{{ .Product }}"
	project          = "{{ .Project }}"
	harbor_migration = "true"
	shipment         = "{{ .Shipment }}"
}

app = "{{ .App }}"
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

func getTfVarsForEnv(data *ecsTerraformShipmentEnvironment) string {

	tf := `# generated by harbor-compose on {{ .GeneratedDate }}

region = "{{ .AwsRegion }}"

aws_profile = "{{ .AwsProfile }}"
	
saml_role = "{{ .AwsSamlRole }}"

secrets_saml_users = []

tags = {
	application      = "{{ .App }}"
	environment      = "{{ .Env }}"
	team             = "{{ .Group }}"
	customer         = "{{ .Property }}"
	contact-email    = "{{ .ContactEmail }}"
	product          = "{{ .Product }}"
	project          = "{{ .Project }}"
	harbor_migration = "true"
	shipment         = "{{ .Shipment }}"
}

app = "{{ .App }}"

environment = "{{ .Env }}"

internal = "{{ if .PublicLB }}false{{ else }}true{{ end }}"

container_name = "{{ .ContainerName }}"

container_port = "{{ .PrimaryPort.Value }}"

{{ if .HTTPPort }}
lb_port = "{{ .HTTPPort.PublicPort }}"
{{ else }}
lb_port = "{{ .HTTPSPort.PublicPort }}"
{{ end }}

lb_protocol = "HTTP"

replicas = "1"

health_check = "{{ .PrimaryPort.Healthcheck }}"

health_check_interval = "{{ .PrimaryPort.HealthcheckInterval }}"

health_check_timeout = "{{ .PrimaryPort.HealthcheckTimeout }}"

health_check_matcher = "200-299"

old_image = "{{ .OldImage }}"

new_image = "{{ .NewImage }}"

{{ if .HTTPSPort }}
https_port = "{{ .HTTPSPort.PublicPort }}"

{{ if eq .HTTPSPort.SslManagementType "acm" }}
certificate_arn = "{{ .HTTPSPort.SslArn }}"
{{ end }}
{{ end }}

{{ if ne .LogzToken "" }}
# logz.io

logz_token = "{{ .LogzToken }}"
{{ end }}

# networking config

vpc = "{{ .AwsVpc }}"

private_subnets = "{{ .AwsPrivateSubnets }}"

public_subnets = "{{ .AwsPublicSubnets }}"

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

func migrateImage(image string, accountID string, region string, shipment string) string {
	//extract the tag
	//e.g,
	//registry.services.dmtio.net/xyz:0.1.0
	//quay.io/turner/xyz:0.1.0
	//to
	//${accountID}.dkr.ecr.${region}.amazonaws.com/${shipment}:0.1.0
	tmp := strings.Split(image, ":")
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s", accountID, region, shipment, tmp[len(tmp)-1])
}
