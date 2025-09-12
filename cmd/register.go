package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new client certificate with the server",
	Long:  "Register a new client certificate with the server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("register called")
	},
}
