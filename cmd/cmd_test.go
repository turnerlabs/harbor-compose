package cmd

import (
	"flag"
	"os"
	"testing"
)

var usernameTest string
var passwordTest string

// Flags for testing

func TestMain(m *testing.M) {

	// username and password are not required but if you do pass them you get more testes ran
	flag.StringVar(&usernameTest, "usernameTest", "", "username for authorization")
	flag.StringVar(&passwordTest, "passwordTest", "", "password for authorization")
	flag.Parse()

	os.Exit(m.Run())
}
