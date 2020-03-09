package main

import tgbotapi "github.com/0x0BSoD/telegram-bot-api"

func torrentKbd(hash string, status int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Details", "open_"+hash),
		),
	)
}

func torrentDeleteKbd(hash string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", "delete-yes_"+hash),
			tgbotapi.NewInlineKeyboardButtonData("Yes(with data)", "delete-yes+data"+hash),
			tgbotapi.NewInlineKeyboardButtonData("Cancel", "delete-no_"+hash),
		),
	)
}

func torrentQueueKbd(hash string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏫", "prior-top_"+hash),
			tgbotapi.NewInlineKeyboardButtonData("🔼", "prior-up_"+hash),
			tgbotapi.NewInlineKeyboardButtonData("🔽", "prior-down"+hash),
			tgbotapi.NewInlineKeyboardButtonData("⏬", "prior-bottom"+hash),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Cancel", "prior-no_"+hash),
		),
	)
}

func torrentDetailKbd(hash string, status int) tgbotapi.InlineKeyboardMarkup {
	if status == 0 {
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Priority", "priority_"+hash),
				tgbotapi.NewInlineKeyboardButtonData("Files", "files_"+hash),
				tgbotapi.NewInlineKeyboardButtonData("Start", "start_"+hash),
				tgbotapi.NewInlineKeyboardButtonData("Delete", "delete_"+hash),
			),
		)
	} else {
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Priority", "priority_"+hash),
				tgbotapi.NewInlineKeyboardButtonData("Files", "files_"+hash),
				tgbotapi.NewInlineKeyboardButtonData("Stop", "stop_"+hash),
				tgbotapi.NewInlineKeyboardButtonData("Delete", "delete_"+hash),
			),
		)
	}
}

var configKbd = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Show full config as JSON", "json"),
	),
)

var mainKbd = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("All torrents"),
		tgbotapi.NewKeyboardButton("Active torrents"),
		tgbotapi.NewKeyboardButton("Not Active torrents"),
	),
)
