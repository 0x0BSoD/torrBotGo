package main

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"log"
)

func handleCommand(upd tgbotapi.Update) {
	ID := upd.Message.Chat.ID
	msg := tgbotapi.NewMessage(ID, "")
	msg.ParseMode = "MarkdownV2"
	var err error
	var text string
	switch upd.Message.Command() {
	case "help", "start":
		text = "Telegram Bot as interface for transmission"
		msg.ReplyMarkup = mainKbd
	case "status":
		text, err = sendStatus()
	case "config":
		text, err = sendConfig()
		msg.ReplyMarkup = configKbd
	default:
		sendError(ID, "I don't know that command")
		return
	}

	msg.Text = text

	if err != nil {
		sendError(ID, err.Error())
	} else {
		if msg.Text != "" {
			if _, err := ctx.Bot.Send(msg); err != nil {
				log.Panic(err)
			}
		}
	}
}
