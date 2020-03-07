package main

import (
	"fmt"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"strings"
)

func handleInline(upd tgbotapi.Update) {
	if upd.CallbackQuery.Data == "" {
		return
	}
	chatID := upd.CallbackQuery.Message.Chat.ID
	messageID := upd.CallbackQuery.Message.MessageID
	var err error
	if strings.Contains(upd.CallbackQuery.Data, "_") {
		request := strings.Split(upd.CallbackQuery.Data, "_")
		switch request[0] {
		case "open":
			err = sendTorrentDetails(request[1], chatID, messageID)
		case "delete":
			err = removeTorrentQuestion(request[1], chatID, messageID)
		case "delete-yes":
			err = removeTorrent(request[1], chatID, messageID, request[0])
		case "delete-yes+data":
			err = removeTorrent(request[1], chatID, messageID, request[0])
		case "delete-no":
			err = removeTorrent(request[1], chatID, messageID, request[0])
		case "files":
			sendTorrentFiles(request[1], chatID)
			return
		case "stop":
			err = stopTorrent(request[1], chatID, messageID)
		case "start":
			err = startTorrent(request[1], chatID, messageID)
		case "pUp":
			fmt.Println("pUp ", request[1])
		case "pDown":
			fmt.Println("pDown ", request[1])
		default:
			sendError(chatID, "I don't know that command")
			return
		}
	}
	//else {
	//	switch upd.CallbackQuery.Data {
	//	case "json":
	//		msg.Text = sendJsonConfig()
	//	default:
	//		sendError(ID, "I don't know that command")
	//		return
	//	}
	//}

	if err != nil {
		sendError(chatID, err.Error())
	}
}
