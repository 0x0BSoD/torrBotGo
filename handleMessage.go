package main

import (
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
)

func handleMessage(upd tgbotapi.Update) {
	var (
		result map[string]string
		err    error
	)

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
		result, err = ctx.Transmisson.SearchTorrent(upd.Message.Text)
	} else {
		switch upd.Message.Text {
		case "All torrents":
			result, err = ctx.Transmisson.GetTorrents(all)
		case "Active torrents":
			result, err = ctx.Transmisson.GetTorrents(active)
		case "Not Active torrents":
			result, err = ctx.Transmisson.GetTorrents(notActive)
		default:
			sendError("I don't know that command. handleMessage")
		}
	}

	if err != nil {
		sendError(err.Error())
	} else {
		for hash, text := range result {
			replyMarkup := torrentKbd(hash)
			if err = sendNewMessage(ctx.chatID, text, replyMarkup); err != nil {
				log.Panic(err)
			}
		}
	}
}
