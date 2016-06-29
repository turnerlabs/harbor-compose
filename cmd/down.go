package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop your application",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO")
	},
}

func init() {
	RootCmd.AddCommand(downCmd)
}
