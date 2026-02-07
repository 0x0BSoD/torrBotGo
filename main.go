/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

torrBotGo - Transmission Telegram Bot
Main entry point for the application.
*/
package main

import (
	"time"

	"github.com/0x0BSoD/torrBotGo/cmd"
)

var (
	version   = "2.0.1"
	buildDate string
)

// main is the entry point of the torrBotGo application.
// It delegates execution to the Cobra CLI framework.
func main() {
	buildDate = time.Now().Format(time.RFC3339)

	cmd.Execute()
}
