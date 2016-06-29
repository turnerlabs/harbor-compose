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

// Verbose determines whether or not verbose output is enabled
var Verbose bool

// File represents the
var File string

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Show more output")
	RootCmd.PersistentFlags().StringVarP(&File, "file", "f", "docker-compose.yml", "Specify an alternate compose file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
}
