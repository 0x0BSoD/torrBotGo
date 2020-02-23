package main

import tgbotapi "github.com/0x0BSoD/telegram-bot-api"

func torrentKbd(hash string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Details", "open_"+hash),
			tgbotapi.NewInlineKeyboardButtonData("Delete", "delete_"+hash),
		),
	)
}

var configKbd = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Show full config as JSON", "cfg_json"),
	),
)

var mainKbd = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("All torrents"),
		tgbotapi.NewKeyboardButton("Active torrents"),
		tgbotapi.NewKeyboardButton("Not Active torrents"),
	),
)
