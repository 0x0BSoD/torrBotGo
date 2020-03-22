package main

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"strconv"
	"strings"
)

func handleMessage(upd tgbotapi.Update) {
	var err error

	if torrentID, err := strconv.ParseInt(upd.Message.Text, 10, 64); err == nil {
		err = sendTorrentDetailsByID(torrentID)
		if err != nil {
			sendError(err.Error())
		}
		return
	}

	if upd.Message.Document != nil {
		err = addTorrentFileQuestion(upd.Message.Document.FileID)
	} else if strings.HasPrefix(upd.Message.Text, "magnet:") {
		err = addTorrentMagnetQuestion(upd.Message.Text)
	} else {
		switch upd.Message.Text {
		case "All torrents":
			err = sendTorrentList(All)
		case "Active torrents":
			err = sendTorrentList(Active)
		case "Not Active torrents":
			err = sendTorrentList(NotActive)
		default:
			sendError("I don't know that command. handleMessage")
		}
	}

	if err != nil {
		sendError(err.Error())
	}
}
