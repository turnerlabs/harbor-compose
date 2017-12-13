package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:    "harbor-compose",
	Short:  "Define and run multi-container Docker apps on Harbor",
	Long:   ``,
	PreRun: preRunHook,
}

func preRunHook(cmd *cobra.Command, args []string) {
	currentCommand = cmd.Name()
	writeMetric(currentCommand)
}

// Version is the version of this app
var Version string

// BuildDate is the date this binary was built
var BuildDate string

// Verbose determines whether or not verbose output is enabled
var Verbose bool

// DockerComposeFile represents the docker-compose.yml file
var DockerComposeFile string

// HarborComposeFile represents the harbor-compose.yml file
var HarborComposeFile string

//currently executing command
var currentCommand string

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string, buildDate string) {
	Version = version
	BuildDate = buildDate

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Show more output")
	RootCmd.PersistentFlags().StringVarP(&DockerComposeFile, "file", "f", "docker-compose.yml", "Specify an alternate docker compose file")
	RootCmd.PersistentFlags().StringVarP(&HarborComposeFile, "harbor-file", "c", "harbor-compose.yml", "Specify an alternate harbor compose file")
}
