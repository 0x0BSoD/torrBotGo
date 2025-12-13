package app

import (
	"fmt"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"

	"github.com/0x0BSoD/torrBotGo/internal/telegram"
)

func (h *handler) handleCommand(update tgbotapi.Update) {
	h.tClient.SetChatID(update.Message.Chat.ID)

	switch update.Message.Command() {
	case "help", "start":
		sendMessageWrapper(h.tClient, nil, telegram.MainKeyboard, "Telegram Bot as interface for transmission")
	case "config":
		config, err := h.trClient.SessionConfig()
		if err != nil {
			h.tClient.SendError(fmt.Sprintf("get config failed, %v", err))
			return
		}
		sendMessageWrapper(h.tClient, telegram.TmplConfig(), telegram.ConfigKbd, config)
	case "status":
		status, err := h.trClient.Status()
		if err != nil {
			h.tClient.SendError(fmt.Sprintf("get status failed, %v", err))
			return
		}
		sendMessageWrapper(h.tClient, telegram.TmplStatus(), nil, status)
	default:
		h.tClient.SendError("I don't know that command.")
		return
	}
}
