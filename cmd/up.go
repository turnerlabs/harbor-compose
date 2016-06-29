package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start your application",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("up called")
	},
}

func init() {
	RootCmd.AddCommand(upCmd)
}
