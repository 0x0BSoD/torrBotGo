package main

import (
	"strings"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
)

func handleInline(upd tgbotapi.Update) {
	if upd.CallbackQuery == nil || upd.CallbackQuery.Data == "" {
		return
	}

	messageID := upd.CallbackQuery.Message.MessageID
	ctx.chatID = upd.CallbackQuery.Message.Chat.ID
	var err error

	if strings.HasPrefix(upd.CallbackQuery.Data, "file+add-") {
		err = ctx.Transmisson.addTorrentFile(upd.CallbackQuery.Data)
	}

	if strings.HasPrefix(upd.CallbackQuery.Data, "add-") {
		err = ctx.Transmisson.addTorrentMagnet(upd.CallbackQuery.Data)
	}

	if strings.Contains(upd.CallbackQuery.Data, "_") {
		t := strings.ReplaceAll(strings.ReplaceAll(upd.CallbackQuery.Message.Text, "`", ""), "\n", "")

		request := strings.Split(upd.CallbackQuery.Data, "_")
		switch request[0] {
		case "open", "update":
			err = ctx.Transmisson.sendTorrentDetails(request[1], messageID, glh.GetMD5Hash(t))
		case "delete":
			err = ctx.Transmisson.removeTorrentQuestion(request[1], messageID)
		case "delete-yes":
			err = ctx.Transmisson.removeTorrent(request[1], messageID, request[0])
		case "delete-yes+data":
			err = ctx.Transmisson.removeTorrent(request[1], messageID, request[0])
		case "delete-no":
			err = ctx.Transmisson.removeTorrent(request[1], messageID, request[0])
		case "files":
			err = ctx.Transmisson.sendTorrentFiles(request[1])
		case "stop":
			err = ctx.Transmisson.stopTorrent(request[1], messageID, glh.GetMD5Hash(t))
		case "start":
			err = ctx.Transmisson.startTorrent(request[1], messageID, glh.GetMD5Hash(t))
		case "priority":
			err = ctx.Transmisson.queueTorrentQuestion(request[1], messageID)
		case "prior-top":
			err = ctx.Transmisson.queueTorrent(request[1], messageID, request[0])
		case "prior-up":
			err = ctx.Transmisson.queueTorrent(request[1], messageID, request[0])
		case "prior-down":
			err = ctx.Transmisson.queueTorrent(request[1], messageID, request[0])
		case "prior-bottom":
			err = ctx.Transmisson.queueTorrent(request[1], messageID, request[0])
		case "prior-no":
			err = ctx.Transmisson.queueTorrent(request[1], messageID, request[0])
		case "json":
			err = ctx.Transmisson.sendJSONConfig()
		default:
			sendError("I don't know that command, handleInline")
			return
		}
	}

	if err != nil {
		sendError(err.Error())
	}
}
