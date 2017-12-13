package cmd

import (
	"flag"
	"os"
	"testing"
)

// Flags for testing
var (
	integrationTest = flag.Bool("integrationTest", false, "run integration tests")
	usernameTest    = flag.String("usernameTest", "", "username for authorization")
	passwordTest    = flag.String("passwordTest", "", "password for authorization")
)

func TestMain(m *testing.M) {
	flag.Parse()

	//allow integration test envvars to override flags
	if temp := os.Getenv("HC_INTEGRATION_TEST_USERNAME"); temp != "" {
		usernameTest = &temp
	}

	if temp := os.Getenv("HC_INTEGRATION_TEST_PASSWORD"); temp != "" {
		passwordTest = &temp
	}

	//run tests
	os.Exit(m.Run())
}
