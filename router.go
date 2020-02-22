package main

import (
	"fmt"
	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"log"
)

func parseUpdate(upd tgbotapi.Update) {
	msg := tgbotapi.NewMessage(-1, "")
	msg.ParseMode = "MarkdownV2"

	if upd.Message == nil {
		if upd.CallbackQuery.Data != "" {
			msg = tgbotapi.NewMessage(upd.CallbackQuery.Message.Chat.ID, "")
			msg.ParseMode = "MarkdownV2"
			switch upd.CallbackQuery.Data {
			case "cfg_json":
				msg.Text = sendJsonConfig()
			default:
				msg := tgbotapi.NewVideoUpload(upd.CallbackQuery.Message.Chat.ID, "error.mp4")
				if _, err := ctx.Bot.Send(msg); err != nil {
					log.Panic(err)
				}
				return
			}
		}
	} else {
		msg = tgbotapi.NewMessage(upd.Message.Chat.ID, "")
		msg.ParseMode = "MarkdownV2"

		if !upd.Message.IsCommand() {
			fmt.Println("Sample message type:")
			_ = glh.PrettyPrint(upd)
			msg.Text = "ok"
		} else {
			switch upd.Message.Command() {
			case "help":
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

	if _, err := ctx.Bot.Send(msg); err != nil {
		log.Panic(err)
	}
}
