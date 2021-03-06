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
		err = addTorrentFile(upd.CallbackQuery.Data)
	}

	if strings.HasPrefix(upd.CallbackQuery.Data, "add-") {
		err = addTorrentMagnet(upd.CallbackQuery.Data)
	}

	if strings.Contains(upd.CallbackQuery.Data, "_") {
		t := strings.ReplaceAll(strings.ReplaceAll(upd.CallbackQuery.Message.Text, "`", ""), "\n", "")

		request := strings.Split(upd.CallbackQuery.Data, "_")
		switch request[0] {
		case "open", "update":
			err = sendTorrentDetails(request[1], messageID, glh.GetMD5Hash(t))
		case "delete":
			err = removeTorrentQuestion(request[1], messageID)
		case "delete-yes":
			err = removeTorrent(request[1], messageID, request[0])
		case "delete-yes+data":
			err = removeTorrent(request[1], messageID, request[0])
		case "delete-no":
			err = removeTorrent(request[1], messageID, request[0])
		case "files":
			err = sendTorrentFiles(request[1])
		case "stop":
			err = stopTorrent(request[1], messageID, glh.GetMD5Hash(t))
		case "start":
			err = startTorrent(request[1], messageID, glh.GetMD5Hash(t))
		case "priority":
			err = queueTorrentQuestion(request[1], messageID)
		case "prior-top":
			err = queueTorrent(request[1], messageID, request[0])
		case "prior-up":
			err = queueTorrent(request[1], messageID, request[0])
		case "prior-down":
			err = queueTorrent(request[1], messageID, request[0])
		case "prior-bottom":
			err = queueTorrent(request[1], messageID, request[0])
		case "prior-no":
			err = queueTorrent(request[1], messageID, request[0])
		case "json":
			err = sendJSONConfig()
		default:
			sendError("I don't know that command, handleInline")
			return
		}
	}

	if err != nil {
		sendError(err.Error())
	}
}
