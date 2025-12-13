package app

import (
	"fmt"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"

	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	"github.com/0x0BSoD/torrBotGo/internal/transmission"
)

func handleCommand(update tgbotapi.Update, tClient *telegram.Client, trClient *transmission.Client) {
	tClient.SetChatID(update.Message.Chat.ID)

	switch update.Message.Command() {
	case "help", "start":
		sendMessageWrapper(tClient, nil, telegram.MainKeyboard, "Telegram Bot as interface for transmission")
	case "config":
		config, err := trClient.SessionConfig()
		if err != nil {
			tClient.SendError(fmt.Sprintf("get config failed, %v", err))
			return
		}
		sendMessageWrapper(tClient, telegram.TmplConfig(), telegram.ConfigKbd, config)
	case "status":
		status, err := trClient.Status()
		if err != nil {
			tClient.SendError(fmt.Sprintf("get status failed, %v", err))
			return
		}
		sendMessageWrapper(tClient, telegram.TmplStatus(), nil, status)
	default:
		tClient.SendError("I don't know that command.")
		return
	}
}
