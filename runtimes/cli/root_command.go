package main

import "github.com/spf13/cobra"

var rootCommand = &cobra.Command{
	Use:   "notegic",
	Short: "Notegic development and verification commands.",
}

func Execute() error {
	return rootCommand.Execute()
}
