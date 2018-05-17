package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateTerraformBackend(t *testing.T) {

	tf := `
terraform {
	backend "s3" {
		region  = "us-east-1"
		profile = ""
		bucket  = ""
		key     = "dev.terraform.tfstate"
	}
}

provider "aws" {
	region  = "${var.region}"
	profile = "${var.aws_profile}"
}

# output

output "lb_dns" {
	value = "${aws_alb.main.dns_name}"
}

output "status" {
	value = "fargate --cluster ${var.app}-${var.environment} service info ${var.app}-${var.environment}"
}	
`

	data := ecsTerraformShipmentEnvironment{
		Shipment:       "my-shipment",
		Env:            "dev",
		AwsAccountName: "my-account",
		AwsRole:        "devops",
	}

	result := updateTerraformBackend(tf, &data)
	t.Log(result)
	assert.Contains(t, result, `profile = "my-account:my-account-devops"`)
	assert.Contains(t, result, `bucket  = "tf-state-my-shipment"`)
}
