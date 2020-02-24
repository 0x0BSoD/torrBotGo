package main

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"log"
)

func handleCommand(upd tgbotapi.Update) {
	ID := upd.Message.Chat.ID
	msg := tgbotapi.NewMessage(ID, "")
	msg.ParseMode = "MarkdownV2"

	switch upd.Message.Command() {
	case "help", "start":
		msg.Text = "Telegram Bot as interface for transmission"
		msg.ReplyMarkup = mainKbd
	case "status":
		msg.Text = sendStatus()
	case "config":
		msg.Text = sendConfig()
		msg.ReplyMarkup = configKbd
	default:
		sendError(ID, "I don't know that command")
		return
	}

	if msg.Text != "" {
		if _, err := ctx.Bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}
