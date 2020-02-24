package main

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"log"
)

func handleMessage(upd tgbotapi.Update) {
	ID := upd.Message.Chat.ID
	msg := tgbotapi.NewMessage(ID, "")
	msg.ParseMode = "MarkdownV2"

	switch upd.Message.Text {
	case "All torrents":
		sendTorrentList(ID, All)
	case "Active torrents":
		sendTorrentList(ID, Active)
	case "Not Active torrents":
		sendTorrentList(ID, NotActive)
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
