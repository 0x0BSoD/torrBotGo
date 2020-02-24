package main

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"log"
)

func sendError(id int64, text string) {
	msg := tgbotapi.NewVideoUpload(id, "error.mp4")
	msg.Caption = text
	if _, err := ctx.Bot.Send(msg); err != nil {
		log.Panic(err)
	}
}
