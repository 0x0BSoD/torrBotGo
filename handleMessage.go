package main

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
)

func handleMessage(upd tgbotapi.Update) {
	ID := upd.Message.Chat.ID
	var err error
	switch upd.Message.Text {
	case "All torrents":
		err = sendTorrentList(ID, All)
	case "Active torrents":
		err = sendTorrentList(ID, Active)
	case "Not Active torrents":
		err = sendTorrentList(ID, NotActive)
	default:
		sendError(ID, "I don't know that command")
	}

	if err != nil {
		sendError(ID, err.Error())
	}
}
