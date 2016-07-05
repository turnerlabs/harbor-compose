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
	Run:   up,
}

func init() {
	RootCmd.AddCommand(upCmd)
}

func up(cmd *cobra.Command, args []string) {
	fmt.Println("TODO")
}
