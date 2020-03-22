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
		ctx.chatID = upd.Message.Chat.ID
		text = "Telegram Bot as interface for transmission"
		msg.ReplyMarkup = mainKbd
	case "status":
		text, err = sendStatus()
	case "config":
		text, err = sendConfig()
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
