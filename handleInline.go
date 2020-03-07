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
	ID := upd.CallbackQuery.Message.Chat.ID
	var err error
	if strings.Contains(upd.CallbackQuery.Data, "_") {
		request := strings.Split(upd.CallbackQuery.Data, "_")
		switch request[0] {
		case "open":
			err = sendTorrentDetails(request[1], ID, upd.CallbackQuery.Message.MessageID)
		//case "delete":
		//	fmt.Println("delete ", request[1])
		//	msg.Text = removeTorrent()
		case "files":
			sendTorrentFiles(request[1], ID)
			return
		case "stop":
			err = stopTorrent(request[1], ID, upd.CallbackQuery.Message.MessageID)
		case "start":
			err = startTorrent(request[1], ID, upd.CallbackQuery.Message.MessageID)
		case "pUp":
			fmt.Println("pUp ", request[1])
		case "pDown":
			fmt.Println("pDown ", request[1])
		default:
			sendError(ID, "I don't know that command")
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
		sendError(ID, err.Error())
	}
}
