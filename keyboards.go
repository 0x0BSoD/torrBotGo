package main

import tgbotapi "github.com/0x0BSoD/telegram-bot-api"

var configKbd = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Show full config as JSON", "cfg_json"),
	),
)

var mainKbd = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Torrents"),
	),
)
