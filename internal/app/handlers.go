package app

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"

	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	"github.com/0x0BSoD/torrBotGo/internal/transmission"
)

func handleCommand(update tgbotapi.Update, tClient *telegram.Client, trClient *transmission.Client, logger *zap.Logger) {
	chatID := update.Message.Chat.ID

	switch update.Message.Command() {
	case "help", "start":
		if err := tClient.SendMessage(chatID, "Telegram Bot as interface for transmission", telegram.MainKeyboard); err != nil {
			tClient.SendError(chatID, fmt.Sprintf("send help failed, %v", err))
			return
		}
	case "config":
		config, err := trClient.SessionConfig()
		if err != nil {
			tClient.SendError(chatID, fmt.Sprintf("get config failed, %v", err))
			return
		}

		var buf bytes.Buffer
		if err := telegram.TmplConfig().Execute(&buf, config); err != nil {
			tClient.SendError(chatID, fmt.Sprintf("tmpl config failed, %v", err))
			return
		}

		if err := tClient.SendMessage(chatID, buf.String(), telegram.ConfigKbd); err != nil {
			tClient.SendError(chatID, fmt.Sprintf("send config failed, %v", err))
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

func handleInline(update tgbotapi.Update, tClient *telegram.Client, trClient *transmission.Client, logger *zap.Logger) {
	if update.CallbackQuery == nil || update.CallbackQuery.Data == "" {
		return
	}

	// messageID := update.CallbackQuery.Message.MessageID
	chatID := update.CallbackQuery.Message.Chat.ID

	if strings.Contains(update.CallbackQuery.Data, "_") {
		// tHash := glh.GetMD5Hash(strings.ReplaceAll(strings.ReplaceAll(update.CallbackQuery.Message.Text, "`", ""), "\n", ""))
		request := strings.Split(update.CallbackQuery.Data, "_")
		switch request[0] {
		case "json":
			config, err := trClient.SessionJSONConfig()
			if err != nil {
				tClient.SendError(chatID, fmt.Sprintf("get config failed, %v", err))
				return
			}

			if err := tClient.SendMessage(chatID, config, nil); err != nil {
				tClient.SendError(chatID, fmt.Sprintf("send config failed, %v", err))
				return
			}
		}
	}
}

func handleMessage(update tgbotapi.Update, tClient *telegram.Client, trClient *transmission.Client, logger *zap.Logger) {
	chatID := update.Message.Chat.ID
	torrents, err := trClient.Torrents(update.Message.Text)
	if err != nil {
		if errors.Is(err, transmission.ErrorFilterNotFound) {
			tClient.SendError(chatID, "I don't know that command. handleMessage")
			return
		}
		tClient.SendError(chatID, fmt.Sprintf("get torrents failed, %v", err))
		return
	}

	if len(torrents) == 0 {
		if err := tClient.SendMessage(chatID, "Noting to show", nil); err != nil {
			tClient.SendError(chatID, fmt.Sprintf("send failed, %v", err))
			return
		}
	}

	for hash, torrent := range torrents {
		text, err := renderTorrent(torrent)
		if err != nil {
			tClient.SendError(chatID, fmt.Sprintf("render torrent template failed, %v", err))
			return
		}
		replyMarkup := telegram.TorrentKbd(hash)

		if err := tClient.SendMessage(chatID, text, replyMarkup); err != nil {
			tClient.SendError(chatID, fmt.Sprintf("send torrent item failed, %v", err))
			return
		}
	}
}
