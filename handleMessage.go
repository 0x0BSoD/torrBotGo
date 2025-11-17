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

	// If we got message that contains integer value
	if torrentID, err := strconv.ParseInt(upd.Message.Text, 10, 64); err == nil {
		hash, text, err := ctx.Transmisson.TorrentDetailsByID(torrentID)
		if err != nil {
			sendError(err.Error())
		}
		replyMarkup := torrentDetailKbd(hash, TORRENT.Status)
		if err = sendNewMessage(ctx.chatID, text, replyMarkup); err != nil {
			log.Panic(err)
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
			result, err = ctx.Transmisson.Torrents(all)
		case "Active torrents":
			result, err = ctx.Transmisson.Torrents(active)
		case "Not Active torrents":
			result, err = ctx.Transmisson.Torrents(notActive)
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
