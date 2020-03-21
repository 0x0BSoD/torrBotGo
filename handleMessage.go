package main

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"strconv"
	"strings"
)

func handleMessage(upd tgbotapi.Update) {
	ID := upd.Message.Chat.ID
	var err error

	if torrentID, err := strconv.ParseInt(upd.Message.Text, 10, 64); err == nil {
		err = sendTorrentDetailsByID(torrentID, ctx.chatID)
		if err != nil {
			sendError(ID, err.Error())
		}
		return
	}

	if upd.Message.Document != nil {
		err = addTorrentFileQuestion(ID, upd.Message.Document.FileID)
	} else if strings.HasPrefix(upd.Message.Text, "magnet:") {
		err = addTorrentMagnetQuestion(ID, upd.Message.Text)
	} else {
		switch upd.Message.Text {
		case "All torrents":
			err = sendTorrentList(ID, All)
		case "Active torrents":
			err = sendTorrentList(ID, Active)
		case "Not Active torrents":
			err = sendTorrentList(ID, NotActive)
		default:
			sendError(ID, "I don't know that command. handleMessage")
		}
	}

	if err != nil {
		sendError(ID, err.Error())
	}
}
