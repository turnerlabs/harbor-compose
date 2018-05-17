package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrateImage_Internal(t *testing.T) {

	result := migrateImage("registry.services.dmtio.net/xyz:0.1.0", "123456789012", "us-east-1", "my-service")
	t.Log(result)

	assert.Equal(t, "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-service:0.1.0", result)
}

func TestMigrateImage_Quay(t *testing.T) {
	result := migrateImage("quay.io/turner/xyz:0.1.0", "123456789012", "us-east-1", "my-service")
	t.Log(result)

	assert.Equal(t, "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-service:0.1.0", result)
}
