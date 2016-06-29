package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View output from containers",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO")
	},
}

func init() {
	RootCmd.AddCommand(logsCmd)
}
