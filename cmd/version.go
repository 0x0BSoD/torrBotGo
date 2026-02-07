// Package cmd provides the command-line interface for torrBotGo.
// It uses Cobra CLI framework to define and handle commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information variables
var (
	Version   = "dev"
	GitCommit = "dev"
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print the version number and build date of torrBotGo`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("torrBotGo version %s\n", Version)
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Printf("Built on: %s\n", BuildDate)
	},
}

// SetVersion sets the version, git commit, and build date for the application
func SetVersion(version, gitCommit, buildDate string) {
	Version = version
	GitCommit = gitCommit
	BuildDate = buildDate

	// Update root command's Long description to include version info
	rootCmd.Long = `Transmission Telegram Bot

Version: ` + Version + `
Git commit: ` + GitCommit + `
Built: ` + BuildDate
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
