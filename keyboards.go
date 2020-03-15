package main

import tgbotapi "github.com/0x0BSoD/telegram-bot-api"

func torrentAddKbd(byFile bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var btns []tgbotapi.InlineKeyboardButton
	count := 0
	prefix := ""
	if byFile {
		prefix = "file+"
	}

	for name, path := range ctx.Categories {
		if count == 3 {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(btns...))
			btns = []tgbotapi.InlineKeyboardButton{}
			count = 0
		}
		btns = append(btns, tgbotapi.NewInlineKeyboardButtonData(name, prefix+"add-"+path))
		count++
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btns...))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Cancel", prefix+"add-no")))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func torrentKbd(hash string) tgbotapi.InlineKeyboardMarkup {
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
				tgbotapi.NewInlineKeyboardButtonData("🔁", "update_"+hash),
			),
		)
	} else {
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Priority", "priority_"+hash),
				tgbotapi.NewInlineKeyboardButtonData("Files", "files_"+hash),
				tgbotapi.NewInlineKeyboardButtonData("Stop", "stop_"+hash),
				tgbotapi.NewInlineKeyboardButtonData("Delete", "delete_"+hash),
				tgbotapi.NewInlineKeyboardButtonData("🔁", "update_"+hash),
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
