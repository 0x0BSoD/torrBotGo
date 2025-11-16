package main

import (
	"log"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
)

func handleCommand(upd tgbotapi.Update) {
	ctx.chatID = upd.Message.Chat.ID

	var (
		err  error
		text string
		kbd  any
	)

	switch upd.Message.Command() {
	case "help", "start":
		text = "Telegram Bot as interface for transmission"
		kbd = mainKbd
	case "status":
		text, err = ctx.Transmisson.GetStatus()
		kbd = nil
	case "config":
		text, err = ctx.Transmisson.GetConfig()
		kbd = &configKbd
	default:
		sendError("I don't know that command.")
		return
	}

	if err != nil {
		sendError(err.Error())
	} else {
		if text != "" {
			if err = sendNewMessage(ctx.chatID, text, kbd); err != nil {
				log.Panic(err)
			}
		}
	}
}
