/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

torrBotGo - Transmission Telegram Bot
Main entry point for the application.
*/
package main

import "github.com/0x0BSoD/torrBotGo/cmd"

// main is the entry point of the torrBotGo application.
// It delegates execution to the Cobra CLI framework.
func main() {
	cmd.Execute()
}
