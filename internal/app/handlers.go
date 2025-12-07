package app

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"

	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	"github.com/0x0BSoD/torrBotGo/internal/transmission"
)

func handleCommand(update tgbotapi.Update, tClient *telegram.Client, trClient *transmission.Client, logger *zap.Logger) {
	tClient.SetChatID(update.Message.Chat.ID)

	switch update.Message.Command() {
	case "help", "start":
		fmt.Println(update.Message.Chat.ID)
		if err := tClient.SendMessage("Telegram Bot as interface for transmission", telegram.MainKeyboard); err != nil {
			tClient.SendError(fmt.Sprintf("send help failed, %v", err))
			return
		}
	case "config":
		config, err := trClient.SessionConfig()
		if err != nil {
			tClient.SendError(fmt.Sprintf("get config failed, %v", err))
			return
		}

		var buf bytes.Buffer
		if err := telegram.TmplConfig().Execute(&buf, config); err != nil {
			tClient.SendError(fmt.Sprintf("tmpl config failed, %v", err))
			return
		}

		if err := tClient.SendMessage(buf.String(), telegram.ConfigKbd); err != nil {
			tClient.SendError(fmt.Sprintf("send config failed, %v", err))
			return
		}
	case "status":
		status, err := trClient.Status()
		if err != nil {
			tClient.SendError(fmt.Sprintf("get status failed, %v", err))
			return
		}

		var buf bytes.Buffer
		if err := telegram.TmplStatus().Execute(&buf, status); err != nil {
			tClient.SendError(fmt.Sprintf("tmpl status failed, %v", err))
			return
		}

		if err := tClient.SendMessage(buf.String(), nil); err != nil {
			tClient.SendError(fmt.Sprintf("send help failed, %v", err))
			return
		}
	default:
		tClient.SendError("I don't know that command.")
		return
	}
}

func handleInline(update tgbotapi.Update, tClient *telegram.Client, trClient *transmission.Client, logger *zap.Logger) {
	if update.CallbackQuery == nil || update.CallbackQuery.Data == "" {
		return
	}

	messageID := update.CallbackQuery.Message.MessageID
	tClient.SetChatID(update.CallbackQuery.Message.Chat.ID)

	if strings.HasPrefix(update.CallbackQuery.Data, "file+add-") {
		text, err := trClient.AddTorrentByFile(update.CallbackQuery.Data)
		if err != nil {
			tClient.SendError(fmt.Sprintf("add torrent by file failed, %v", err))
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
		case "open", "update", "delete-no":
			torrent, err := trClient.TorrentDetails(hash)
			if err != nil {
				tClient.SendError(fmt.Sprintf("get torrent failed, %v", err))
				return
			}

			var buf bytes.Buffer
			if err := telegram.TmplTorrent().Execute(&buf, torrent); err != nil {
				tClient.SendError(fmt.Sprintf("tmpl torrent failed, %v", err))
				return
			}

			oldHash := glh.GetMD5Hash(strings.ReplaceAll(strings.ReplaceAll(update.CallbackQuery.Message.Text, "`", ""), "\n", ""))
			newHash := glh.GetMD5Hash(strings.ReplaceAll(strings.ReplaceAll(buf.String(), "`", ""), "\n", ""))
			if newHash == oldHash {
				return
			}

			replyMarkup := telegram.TorrentDetailKbd(hash, torrent.StatusCode)
			if err := tClient.SendEditedMessage(messageID, buf.String(), &replyMarkup); err != nil {
				tClient.SendError(fmt.Sprintf("send torrent details failed, %v", err))
				return
			}
		case "delete":
			torrent, err := trClient.TorrentDetails(hash)
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
		case "stop":
		case "start":
		case "priority":
		case "prior-top":
		case "prior-up":
		case "prior-down":
		case "prior-bottom":
		case "prior-no":
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

func handleMessage(update tgbotapi.Update, tClient *telegram.Client, trClient *transmission.Client, logger *zap.Logger) {
	tClient.SetChatID(update.Message.Chat.ID)

	if update.Message.Document != nil {
		_url, err := tClient.BotAPI.GetFileDirectURL(update.Message.Document.FileID)
		if err != nil {
			tClient.SendError(fmt.Sprintf("get file URL failed, %v", err))
			return
		}
		title, imgPath, err := trClient.AddTorrentByFileDialog(_url)
		if err != nil {
			tClient.SendError(fmt.Sprintf("add torrent by file failed, %v", err))
			return
		}

		catList := extractKeys(trClient.Categories)
		data := strings.Split(title, "::")
		if len(data) >= 2 {
			suggestedCat := data[0]
			title = data[1]
			if suggestedCat != "noop" {
				catList = []string{
					suggestedCat,
				}
			}
		}
		kbdAdd := telegram.TorrentAddKbd(true, catList)
		if imgPath != "" {
			if err := tClient.SendImagedMessage(title, imgPath, kbdAdd); err != nil {
				tClient.SendError(fmt.Sprintf("send failed, %v", err))
				return
			}
			return
		}
		if err := tClient.SendMessage(title, kbdAdd); err != nil {
			tClient.SendError(fmt.Sprintf("send failed, %v", err))
			return
		}
	} else {
		torrents, err := trClient.Torrents(update.Message.Text)
		if err != nil {
			if errors.Is(err, transmission.ErrorFilterNotFound) {
				tClient.SendError("I don't know that command. handleMessage")
				return
			}
			tClient.SendError(fmt.Sprintf("get torrents failed, %v", err))
			return
		}

		if len(torrents) == 0 {
			if err := tClient.SendMessage("Noting to show", nil); err != nil {
				tClient.SendError(fmt.Sprintf("send failed, %v", err))
				return
			}
		}

		for hash, torrent := range torrents {
			text, err := renderTorrent(torrent)
			if err != nil {
				tClient.SendError(fmt.Sprintf("render torrent template failed, %v", err))
				return
			}
			replyMarkup := telegram.TorrentKbd(hash)

			if err := tClient.SendMessage(text, replyMarkup); err != nil {
				tClient.SendError(fmt.Sprintf("send torrent item failed, %v", err))
				return
			}
		}
	}
}
