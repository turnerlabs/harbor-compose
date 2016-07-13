package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and exit",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%v built on %v\n", Version, BuildDate)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
