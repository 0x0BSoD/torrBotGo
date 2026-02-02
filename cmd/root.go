// Package cmd provides the command-line interface for torrBotGo.
// It uses Cobra CLI framework to define and handle commands.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "torrbot",
	Short: "TorrBot",
	Long:  `Transmission Telegram Bot`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
