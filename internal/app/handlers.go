package app

import (
	"bytes"
	"fmt"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	"github.com/0x0BSoD/torrBotGo/internal/transmission"
	"go.uber.org/zap"
)

func handleCommand(update tgbotapi.Update, tClient *telegram.Client, trClient *transmission.Client, logger *zap.Logger) {
	chatID := update.Message.Chat.ID

	switch update.Message.Command() {
	case "help", "start":
		if err := tClient.SendMessage(chatID, "Telegram Bot as interface for transmission", telegram.MainKeyboard); err != nil {
			tClient.SendError(chatID, fmt.Sprintf("send help failed, %v", err))
			return
		}
	case "status":
		status, err := trClient.Status()
		if err != nil {
			tClient.SendError(chatID, fmt.Sprintf("get status failed, %v", err))
			return
		}

		var buf bytes.Buffer
		if err := telegram.TmplStatus().Execute(&buf, status); err != nil {
			tClient.SendError(chatID, fmt.Sprintf("tmpl status failed, %v", err))
			return
		}

		if err := tClient.SendMessage(chatID, buf.String(), nil); err != nil {
			tClient.SendError(chatID, fmt.Sprintf("send help failed, %v", err))
			return
		}
	default:
		tClient.SendError(chatID, "I don't know that command.")
		return
	}
}
