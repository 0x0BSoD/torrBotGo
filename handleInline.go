package main

import (
	"fmt"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"log"
	"strings"
)

func handleInline(upd tgbotapi.Update) {
	if upd.CallbackQuery.Data == "" {
		return
	}

	ID := upd.CallbackQuery.Message.Chat.ID
	msg := tgbotapi.NewMessage(ID, "")
	msg.ParseMode = "MarkdownV2"

	if strings.Contains(upd.CallbackQuery.Data, "_") {
		request := strings.Split(upd.CallbackQuery.Data, "_")
		switch request[0] {
		case "open":
			msg.Text = sendTorrentDetails(request[1])
		case "delete":
			fmt.Println("delete ", request[1])
		default:
			sendError(ID, "I don't know that command")
			return
		}
	} else {
		switch upd.CallbackQuery.Data {
		case "json":
			msg.Text = sendJsonConfig()
		default:
			sendError(ID, "I don't know that command")
			return
		}
	}

	if msg.Text != "" {
		if _, err := ctx.Bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}
