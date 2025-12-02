package app

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	"go.uber.org/zap"
)

func handleCommand(update tgbotapi.Update, tClient *telegram.Client, logger *zap.Logger) {
	chatID := update.Message.Chat.ID

	switch update.Message.Command() {
	case "help", "start":
		if err := tClient.SendMessage(chatID, "Telegram Bot as interface for transmission", telegram.MainKeyboard); err != nil {
			logger.Sugar().Errorf("send message failed, %v", err)
			return
		}
	default:
		tClient.SendError(chatID, "I don't know that command.")
		return
	}
}
