package telegram

import tgbotapi "github.com/0x0BSoD/telegram-bot-api"

var MainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("All torrents"),
		tgbotapi.NewKeyboardButton("Active torrents"),
		tgbotapi.NewKeyboardButton("Not Active torrents"),
	),
)

var ConfigKbd = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Show full config as JSON", "json_show"),
	),
)

func TorrentKbd(hash string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Details", "open_"+hash),
		),
	)
}

func TorrentAddKbd(byFile bool, categories []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var btns []tgbotapi.InlineKeyboardButton
	count := 0
	prefix := ""
	if byFile {
		prefix = "file+"
	}

	for _, name := range categories {
		if count == 3 {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(btns...))
			btns = []tgbotapi.InlineKeyboardButton{}
			count = 0
		}
		btns = append(btns, tgbotapi.NewInlineKeyboardButtonData(name, prefix+"add-"+name))
		count++
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btns...))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Cancel", prefix+"add-no")))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TorrentDetailKbd(hash string, status int) tgbotapi.InlineKeyboardMarkup {
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
	}

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
