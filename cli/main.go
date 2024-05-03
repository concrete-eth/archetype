package cli

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func logInfo(a ...any) {
	fmt.Println(a...)
}

func logDebug(a ...any) {
	gray.Println(a...)
}

func logWarning(warning string) {
	yellow.Fprint(os.Stderr, "Warning: ")
	fmt.Fprintln(os.Stderr, warning)
}

func logError(err error) {
	fmt.Fprintln(os.Stderr, "Error:")
	red.Fprint(os.Stderr, err)
	fmt.Fprintln(os.Stderr, "\nContext:")
	gray.Fprintln(os.Stderr, string(debug.Stack()))
}

func logFatal(err error) {
	logError(err)
	os.Exit(1)
}

/* CLI */

// NewRootCmd creates the root command for the CLI.
func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{Use: "archetype"}
	AddCodegenCommand(rootCmd)
	AddSnapshotCommand(rootCmd)
	return rootCmd
}

// Execute runs the CLI.
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		logFatal(err)
	}
}
