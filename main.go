/*
Copyright © 2025 0x0BSoD zlodey23@gmail.com

torrBotGo - Transmission Telegram Bot
Main entry point for the application.
*/
package main

import (
	"github.com/0x0BSoD/torrBotGo/cmd"
)

var (
	Version   = "2.0.1"
	GitCommit = "dev"
	BuildDate = "unknown"
)

// main is the entry point of the torrBotGo application.
// It delegates execution to the Cobra CLI framework.
func main() {
	cmd.SetVersion(Version, GitCommit, BuildDate)

	cmd.Execute()
}
