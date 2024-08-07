package cli

import (
	"github.com/spf13/cobra"
)

/* CLI */

// NewRootCmd creates the root command for the CLI.
func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{Use: "archetype"}
	AddCodegenCommand(rootCmd)
	AddSnapshotCommand(rootCmd)
	AddInfoCommand(rootCmd)
	return rootCmd
}

// Execute runs the CLI.
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		logFatal(err)
	}
}

// TODO: is hex address
