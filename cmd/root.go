package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "harbor-compose",
	Short: "Define and run multi-container Docker apps on Harbor",
	Long:  ``,
}

// Version is the version of this app
var Version string

// Verbose determines whether or not verbose output is enabled
var Verbose bool

// DockerComposeFile represents the docker-compose.yml file
var DockerComposeFile string

// HarborComposeFile represents the harbor-compose.yml file
var HarborComposeFile string

// User is the current user
var User string

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	Version = version

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Show more output")
	RootCmd.PersistentFlags().StringVarP(&DockerComposeFile, "file", "f", "docker-compose.yml", "Specify an alternate docker compose file")
	RootCmd.PersistentFlags().StringVarP(&HarborComposeFile, "harbor-file", "c", "harbor-compose.yml", "Specify an alternate harbor compose file")
	RootCmd.PersistentFlags().StringVarP(&User, "user", "u", "", "Runs commands as this user")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
}
