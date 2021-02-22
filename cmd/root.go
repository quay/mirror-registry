package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "quay-installer",
		Short: "A generator for Cobra based Applications",
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
