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
`

	data := ecsTerraformShipmentEnvironment{
		Shipment:   "my-shipment",
		Env:        "dev",
		AwsProfile: "my-profile",
	}

	result := updateTerraformBackend(tf, &data)
	t.Log(result)
	assert.Contains(t, result, `profile = "my-profile"`)
	assert.Contains(t, result, `bucket  = "tf-state-my-shipment"`)
}
