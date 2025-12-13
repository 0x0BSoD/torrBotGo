package app

import (
	"bytes"
	"fmt"
	"strings"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"

	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	"github.com/0x0BSoD/torrBotGo/internal/transmission"
)

func handleInline(update tgbotapi.Update, tClient *telegram.Client, trClient *transmission.Client) {
	if update.CallbackQuery == nil || update.CallbackQuery.Data == "" {
		return
	}

	messageID := update.CallbackQuery.Message.MessageID
	tClient.SetChatID(update.CallbackQuery.Message.Chat.ID)

	if strings.Contains(update.CallbackQuery.Data, "add") {
		var (
			text string
			err  error
		)

		if strings.HasPrefix(update.CallbackQuery.Data, "file+add-") {
			text, err = trClient.AddByFile(update.CallbackQuery.Data)
		} else {
			text, err = trClient.AddByMagent(update.CallbackQuery.Data)
		}
		if err != nil {
			tClient.SendError(fmt.Sprintf("add torrent failed, %v", err))
			return
		}

		if err := tClient.RemoveMessage(messageID); err != nil {
			tClient.SendError(fmt.Sprintf("remove message failed, %v", err))
			return
		}

		if err := tClient.SendMessage(text, nil); err != nil {
			tClient.SendError(fmt.Sprintf("send config failed, %v", err))
			return
		}
	}

	if strings.Contains(update.CallbackQuery.Data, "_") {
		request := strings.Split(update.CallbackQuery.Data, "_")
		hash := request[1]
		switch request[0] {
		case "open", "update", "delete-no", "prior-no":
			torrent, err := trClient.Details(hash)
			if err != nil {
				tClient.SendError(fmt.Sprintf("get torrent failed, %v", err))
				return
			}

			replyMarkup := telegram.TorrentDetailKbd(hash, torrent.StatusCode)

			sendMessageWrapperHash(update.CallbackQuery.Message.Text, tClient, telegram.TmplTorrent(), replyMarkup, torrent)
		case "delete":
			torrent, err := trClient.Details(hash)
			if err != nil {
				tClient.SendError(fmt.Sprintf("get torrent failed, %v", err))
				return
			}

			var buf bytes.Buffer
			if err := telegram.TmplTorrent().Execute(&buf, torrent); err != nil {
				tClient.SendError(fmt.Sprintf("tmpl torrent failed, %v", err))
				return
			}

			replyMarkup := telegram.TorrentDeleteKbd(hash)
			if err := tClient.SendEditedMessage(messageID, buf.String(), &replyMarkup); err != nil {
				tClient.SendError(fmt.Sprintf("send torrent details failed, %v", err))
				return
			}
		case "delete-yes":
			err := trClient.Delete(hash, false)
			if err != nil {
				tClient.SendError(fmt.Sprintf("remove torrent failed, %v", err))
				return
			}

			if err := tClient.SendEditedMessage(messageID, "Removed", nil); err != nil {
				tClient.SendError(fmt.Sprintf("send torrent deleted failed, %v", err))
				return
			}
		case "delete-yes+data":
			err := trClient.Delete(hash, true)
			if err != nil {
				tClient.SendError(fmt.Sprintf("remove torrent and data failed, %v", err))
				return
			}

			if err := tClient.SendEditedMessage(messageID, "Removed", nil); err != nil {
				tClient.SendError(fmt.Sprintf("send torrent deleted failed, %v", err))
				return
			}
		case "files":
			files := trClient.GetFiles(hash)
			for _, f := range files {
				var buf bytes.Buffer
				if err := telegram.TmplTorrentFilesListItem().Execute(&buf, f); err != nil {
					tClient.SendError(fmt.Sprintf("tmpl torrent file item failed, %v", err))
					return
				}

				if err := tClient.SendMessage(buf.String(), nil); err != nil {
					tClient.SendError(fmt.Sprintf("send torrent file item failed, %v", err))
					return
				}
			}
		case "start", "stop":
			torrent, err := trClient.Details(hash)
			if err != nil {
				tClient.SendError(fmt.Sprintf("get torrent failed, %v", err))
				return
			}

			var buf bytes.Buffer
			if err := telegram.TmplTorrent().Execute(&buf, torrent); err != nil {
				tClient.SendError(fmt.Sprintf("tmpl torrent failed, %v", err))
				return
			}

			replyWaitMarkup := telegram.TorrentDetailKbd(hash, -1)
			if err := tClient.SendEditedMessage(messageID, buf.String(), &replyWaitMarkup); err != nil {
				tClient.SendError(fmt.Sprintf("send torrent details failed, %v", err))
				return
			}

			if err := trClient.StartStop(hash, request[0]); err != nil {
				tClient.SendError(fmt.Sprintf("torrent start/stop failed, %v", err))
				return
			}

			replyMarkup := telegram.TorrentDetailKbd(hash, torrent.StatusCode)
			if err := tClient.SendEditedMessage(messageID, buf.String(), &replyMarkup); err != nil {
				tClient.SendError(fmt.Sprintf("send torrent details failed, %v", err))
				return
			}
		case "priority":
			torrent, err := trClient.Details(hash)
			if err != nil {
				tClient.SendError(fmt.Sprintf("get torrent failed, %v", err))
				return
			}

			var buf bytes.Buffer
			if err := telegram.TmplTorrent().Execute(&buf, torrent); err != nil {
				tClient.SendError(fmt.Sprintf("tmpl torrent failed, %v", err))
				return
			}

			replyMarkup := telegram.TorrentQueueKbd(hash)
			if err := tClient.SendEditedMessage(messageID, buf.String(), &replyMarkup); err != nil {
				tClient.SendError(fmt.Sprintf("send torrent details failed, %v", err))
				return
			}
		case "prior-top", "prior-up", "prior-down", "prior-bottom":
			err := trClient.Priority(request[1], request[0])
			if err != nil {
				tClient.SendError(fmt.Sprintf("change torrent priority failed, %v", err))
				return
			}

			torrent, err := trClient.Details(hash)
			if err != nil {
				tClient.SendError(fmt.Sprintf("get torrent failed, %v", err))
				return
			}

			var buf bytes.Buffer
			if err := telegram.TmplTorrent().Execute(&buf, torrent); err != nil {
				tClient.SendError(fmt.Sprintf("tmpl torrent failed, %v", err))
				return
			}

			replyMarkup := telegram.TorrentDetailKbd(hash, torrent.StatusCode)
			if err := tClient.SendEditedMessage(messageID, buf.String(), &replyMarkup); err != nil {
				tClient.SendError(fmt.Sprintf("send torrent details failed, %v", err))
				return
			}
		case "json":
			config, err := trClient.SessionJSONConfig()
			if err != nil {
				tClient.SendError(fmt.Sprintf("get config failed, %v", err))
				return
			}

			if err := tClient.SendMessage(config, nil); err != nil {
				tClient.SendError(fmt.Sprintf("send config failed, %v", err))
				return
			}
		default:
			tClient.SendError("I don't know that command. handleInline")
			return
		}
	}
}
