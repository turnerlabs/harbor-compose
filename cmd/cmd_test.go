package cmd

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var username string
var password string

// Flags for testing

func TestMain(m *testing.M) {
	flag.StringVar(&username, "username", "", "username for authorization")
	flag.StringVar(&password, "password", "", "password for authorization")
	flag.Parse()

	if len(username) == 0 {
		fmt.Println("Missing username.")
		os.Exit(0)
	}

	if len(password) == 0 {
		fmt.Println("Missing password.")
		os.Exit(0)
	}

	os.Exit(m.Run())
}
