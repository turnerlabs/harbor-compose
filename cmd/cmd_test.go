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
	os.Exit(m.Run())
}
