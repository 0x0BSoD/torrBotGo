package main

import (
	"log"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
)

func handleCommand(upd tgbotapi.Update) {
	ID := upd.Message.Chat.ID
	ctx.chatID = ID
	msg := tgbotapi.NewMessage(ID, "")
	msg.ParseMode = "MarkdownV2"
	var err error
	var text string

	switch upd.Message.Command() {
	case "help", "start":
		text = "Telegram Bot as interface for transmission"
		msg.ReplyMarkup = mainKbd
	case "status":
		text, err = ctx.Transmisson.sendStatus()
	case "config":
		text, err = ctx.Transmisson.sendConfig()
		msg.ReplyMarkup = configKbd
	default:
		sendError("I don't know that command. handleCommand")
		return
	}

	msg.Text = text

	if err != nil {
		sendError(err.Error())
	} else {
		if msg.Text != "" {
			if _, err := ctx.Bot.Send(msg); err != nil {
				log.Panic(err)
			}
		}
	}
}
