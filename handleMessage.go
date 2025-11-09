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
		err = ctx.Transmisson.sendTorrentDetailsByID(torrentID)
		if err != nil {
			sendError(err.Error())
		}
		return
	}

	if upd.Message.Document != nil {
		err = ctx.Transmisson.addTorrentFileQuestion(upd.Message.Document.FileID, upd.Message.MessageID)
	} else if strings.HasPrefix(upd.Message.Text, "magnet:") {
		err = ctx.Transmisson.addTorrentMagnetQuestion(upd.Message.Text, upd.Message.MessageID)
	} else if strings.HasPrefix(upd.Message.Text, "t:") {
		err = ctx.Transmisson.searchTorrent(upd.Message.Text)
	} else {
		switch upd.Message.Text {
		case "All torrents":
			err = ctx.Transmisson.sendTorrentList(all)
		case "Active torrents":
			err = ctx.Transmisson.sendTorrentList(active)
		case "Not Active torrents":
			err = ctx.Transmisson.sendTorrentList(notActive)
		default:
			sendError("I don't know that command. handleMessage")
		}
	}

	if err != nil {
		sendError(err.Error())
	}
}
