package main

import (
	"strconv"
	"strings"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
)

func handleMessage(upd tgbotapi.Update) {
	var err error

	ctx.chatID = upd.Message.Chat.ID

	if torrentID, err := strconv.ParseInt(upd.Message.Text, 10, 64); err == nil {
		err = sendTorrentDetailsByID(torrentID)
		if err != nil {
			sendError(err.Error())
		}
		return
	}

	if upd.Message.Document != nil {
		err = addTorrentFileQuestion(upd.Message.Document.FileID, upd.Message.MessageID)
	} else if strings.HasPrefix(upd.Message.Text, "magnet:") {
		err = addTorrentMagnetQuestion(upd.Message.Text, upd.Message.MessageID)
	} else {
		switch upd.Message.Text {
		case "All torrents":
			err = sendTorrentList(all)
		case "Active torrents":
			err = sendTorrentList(active)
		case "Not Active torrents":
			err = sendTorrentList(notActive)
		default:
			sendError("I don't know that command. handleMessage")
		}
	}

	if err != nil {
		sendError(err.Error())
	}
}
