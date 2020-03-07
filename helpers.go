package main

import (
	"fmt"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"log"
)

func sendError(id int64, text string) {
	msg := tgbotapi.NewVideoUpload(id, "error.mp4")
	msg.Caption = text
	if _, err := ctx.Bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func parseStatus(s int) (string, string) {
	var icon string
	var status string

	switch s {
	case 0:
		icon = "⏹️️"
		status = "Stopped"
	case 1:
		icon = "▶️️"
		status = "Queued to check files"
	case 2:
		icon = "▶️"
		status = "Checking files"
	case 3:
		icon = "▶️️"
		status = "Queued to download"
	case 4:
		icon = "▶️"
		status = "Downloading"
	case 5:
		icon = "▶️️"
		status = "'Queued to seed"
	default:
		icon = "▶️️"
		status = "Seeding"
	}

	return icon, status
}

func sendNewMessage(chatID int64, text string, replyMarkup *tgbotapi.InlineKeyboardMarkup) error {
	if text == "" {
		return fmt.Errorf("message cannot be empty")
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "MarkdownV2"

	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup
	}

	if _, err := ctx.Bot.Send(msg); err != nil {
		return err
	}

	return nil
}

func sendEditedMessage(chatID int64, messageID int, text string, replyMarkup *tgbotapi.InlineKeyboardMarkup) error {
	if text == "" {
		return fmt.Errorf("message cannot be empty")
	}

	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "MarkdownV2"

	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup
	}

	if _, err := ctx.Bot.Send(msg); err != nil {
		return err
	}

	return nil
}
