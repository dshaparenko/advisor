package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "advisor",
	Short: "Advisor",
	Long:  `Advisor is a tool for providing advice on k8s requests and limits`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Advisor")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
