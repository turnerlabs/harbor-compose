package cmd

import (
	"flag"
	"os"
	"testing"
)

var username string
var password string

// Flags for testing

func TestMain(m *testing.M) {
	// username and password are not required but if you do pass them you get more testes ran
	flag.StringVar(&username, "username", "", "username for authorization")
	flag.StringVar(&password, "password", "", "password for authorization")
	flag.Parse()

	os.Exit(m.Run())
}
