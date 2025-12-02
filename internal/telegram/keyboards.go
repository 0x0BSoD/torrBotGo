package telegram

import tgbotapi "github.com/0x0BSoD/telegram-bot-api"

var MainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("All torrents"),
		tgbotapi.NewKeyboardButton("Active torrents"),
		tgbotapi.NewKeyboardButton("Not Active torrents"),
	),
)
