	package cmd

	import (
	  "github.com/spf13/cobra"
	)

	var rootCmd = &cobra.Command{
	  Use:   "sncli",
	  Short: "ServiceNow CLI Tool",
	  Long:  "A CLI tool to interact with ServiceNow instances",
	}

	func Execute() error {
	  return rootCmd.Execute()
	}

	func init() {
	  rootCmd.AddCommand(connectCmd)
	}

