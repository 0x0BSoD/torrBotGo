package main

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
)

func parseUpdate(upd tgbotapi.Update) {
	if upd.Message == nil {
		handleInline(upd)
		return
	}

	if upd.Message.IsCommand() {
		handleCommand(upd)
		return
	}

	handleMessage(upd)
}
