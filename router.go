package main

import (
	"fmt"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"log"
	"strings"
)

func parseUpdate(upd tgbotapi.Update) {
	msg := tgbotapi.NewMessage(-1, "")
	msg.ParseMode = "MarkdownV2"

	if upd.Message == nil {
		if upd.CallbackQuery.Data != "" {
			msg = tgbotapi.NewMessage(upd.CallbackQuery.Message.Chat.ID, "")
			msg.ParseMode = "MarkdownV2"
			if strings.Contains(upd.CallbackQuery.Data, "_") {
				request := strings.Split(upd.CallbackQuery.Data, "_")
				switch request[0] {
				case "open":
					msg.Text = sendTorrentDetails(request[1])
				case "delete":
					fmt.Println("delete ", request[1])
				default:
					msg := tgbotapi.NewVideoUpload(upd.CallbackQuery.Message.Chat.ID, "error.mp4")
					if _, err := ctx.Bot.Send(msg); err != nil {
						log.Panic(err)
					}
					return
				}
			} else {
				switch upd.CallbackQuery.Data {
				case "json":
					msg.Text = sendJsonConfig()
				default:
					msg := tgbotapi.NewVideoUpload(upd.CallbackQuery.Message.Chat.ID, "error.mp4")
					if _, err := ctx.Bot.Send(msg); err != nil {
						log.Panic(err)
					}
					return
				}
			}
		}
	} else {
		msg = tgbotapi.NewMessage(upd.Message.Chat.ID, "")
		msg.ParseMode = "MarkdownV2"

		if !upd.Message.IsCommand() {
			switch upd.Message.Text {
			case "All torrents":
				sendTorrentList(upd.Message.Chat.ID, All)
			case "Active torrents":
				sendTorrentList(upd.Message.Chat.ID, Active)
			case "Not Active torrents":
				sendTorrentList(upd.Message.Chat.ID, NotActive)
			default:
				msg := tgbotapi.NewVideoUpload(upd.Message.Chat.ID, "error.mp4")
				if _, err := ctx.Bot.Send(msg); err != nil {
					log.Panic(err)
				}
				return
			}
		} else {
			switch upd.Message.Command() {
			case "help", "start":
				msg.Text = "Telegram Bot as interface for transmission"
				msg.ReplyMarkup = mainKbd
			case "status":
				msg.Text = sendStatus()
			case "config":
				msg.Text = sendConfig()
				msg.ReplyMarkup = configKbd
			default:
				msg.Text = "I don't know that command"
			}
		}
	}

	if msg.Text != "" {
		if _, err := ctx.Bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}
