package app

import (
	"bytes"
	"fmt"
	"strings"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"

	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	"github.com/0x0BSoD/torrBotGo/internal/transmission"
)

func (h *handler) handleInline(update tgbotapi.Update) {
	if update.CallbackQuery == nil || update.CallbackQuery.Data == "" {
		return
	}

	messageID := update.CallbackQuery.Message.MessageID
	h.tClient.SetChatID(update.CallbackQuery.Message.Chat.ID)

	if strings.Contains(update.CallbackQuery.Data, "add") {
		addTorrent(update.CallbackQuery.Data, messageID, h.tClient, h.trClient)
	}

	if strings.Contains(update.CallbackQuery.Data, "_") {
		request := strings.Split(update.CallbackQuery.Data, "_")
		hash := request[1]
		switch request[0] {
		case "open", "update", "delete-no", "prior-no":
			torrent, err := h.trClient.Details(hash)
			if err != nil {
				h.tClient.SendError(fmt.Sprintf("get torrent failed, %v", err))
				return
			}

			replyMarkup := telegram.TorrentDetailKbd(hash, torrent.StatusCode)

			editMessageWrapperHash(messageID, update.CallbackQuery.Message.Text, h.tClient, telegram.TmplTorrent(), &replyMarkup, torrent)
		case "delete":
			torrent, err := h.trClient.Details(hash)
			if err != nil {
				h.tClient.SendError(fmt.Sprintf("get torrent failed, %v", err))
				return
			}

			replyMarkup := telegram.TorrentDeleteKbd(hash)

			editMessageWrapperHash(messageID, update.CallbackQuery.Message.Text, h.tClient, telegram.TmplTorrent(), replyMarkup, torrent)
		case "delete-yes", "delete-yes+data":
			err := h.trClient.Delete(hash, strings.HasSuffix(request[0], "data"))
			if err != nil {
				h.tClient.SendError(fmt.Sprintf("remove torrent failed, %v", err))
				return
			}

			if err := h.tClient.SendEditedMessage(messageID, "Removed", nil); err != nil {
				h.tClient.SendError(fmt.Sprintf("send torrent deleted failed, %v", err))
				return
			}
		case "files":
			var result struct {
				Files []transmission.TorrentFilesItem
			}
			result.Files = h.trClient.GetFiles(hash)
			var buf bytes.Buffer
			if err := telegram.TmplTorrentFilesListItem().Execute(&buf, result); err != nil {
				h.tClient.SendError(fmt.Sprintf("tmpl torrent file item failed, %v", err))
				return
			}

			if err := h.tClient.SendMessage(buf.String(), nil); err != nil {
				h.tClient.SendError(fmt.Sprintf("send torrent file item failed, %v", err))
				return
			}
		case "start", "stop":
			torrent, err := h.trClient.Details(hash)
			if err != nil {
				h.tClient.SendError(fmt.Sprintf("get torrent failed, %v", err))
				return
			}

			var buf bytes.Buffer
			if err := telegram.TmplTorrent().Execute(&buf, torrent); err != nil {
				h.tClient.SendError(fmt.Sprintf("tmpl torrent failed, %v", err))
				return
			}

			replyWaitMarkup := telegram.TorrentDetailKbd(hash, -1)
			if err := h.tClient.SendEditedMessage(messageID, buf.String(), &replyWaitMarkup); err != nil {
				h.tClient.SendError(fmt.Sprintf("send torrent details failed, %v", err))
				return
			}

			if err := h.trClient.StartStop(hash, request[0]); err != nil {
				h.tClient.SendError(fmt.Sprintf("torrent start/stop failed, %v", err))
				return
			}

			replyMarkup := telegram.TorrentDetailKbd(hash, torrent.StatusCode)
			if err := h.tClient.SendEditedMessage(messageID, buf.String(), &replyMarkup); err != nil {
				h.tClient.SendError(fmt.Sprintf("send torrent details failed, %v", err))
				return
			}
		case "priority":
			torrent, err := h.trClient.Details(hash)
			if err != nil {
				h.tClient.SendError(fmt.Sprintf("get torrent failed, %v", err))
				return
			}

			var buf bytes.Buffer
			if err := telegram.TmplTorrent().Execute(&buf, torrent); err != nil {
				h.tClient.SendError(fmt.Sprintf("tmpl torrent failed, %v", err))
				return
			}

			replyMarkup := telegram.TorrentQueueKbd(hash)
			if err := h.tClient.SendEditedMessage(messageID, buf.String(), &replyMarkup); err != nil {
				h.tClient.SendError(fmt.Sprintf("send torrent details failed, %v", err))
				return
			}
		case "prior-top", "prior-up", "prior-down", "prior-bottom":
			err := h.trClient.Priority(request[1], request[0])
			if err != nil {
				h.tClient.SendError(fmt.Sprintf("change torrent priority failed, %v", err))
				return
			}

			torrent, err := h.trClient.Details(hash)
			if err != nil {
				h.tClient.SendError(fmt.Sprintf("get torrent failed, %v", err))
				return
			}

			var buf bytes.Buffer
			if err := telegram.TmplTorrent().Execute(&buf, torrent); err != nil {
				h.tClient.SendError(fmt.Sprintf("tmpl torrent failed, %v", err))
				return
			}

			replyMarkup := telegram.TorrentDetailKbd(hash, torrent.StatusCode)
			if err := h.tClient.SendEditedMessage(messageID, buf.String(), &replyMarkup); err != nil {
				h.tClient.SendError(fmt.Sprintf("send torrent details failed, %v", err))
				return
			}
		case "json":
			config, err := h.trClient.SessionJSONConfig()
			if err != nil {
				h.tClient.SendError(fmt.Sprintf("get config failed, %v", err))
				return
			}

			if err := h.tClient.SendMessage(config, nil); err != nil {
				h.tClient.SendError(fmt.Sprintf("send config failed, %v", err))
				return
			}
		default:
			h.tClient.SendError("I don't know that command. handleInline")
			return
		}
	}
}
