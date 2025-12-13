package app

import (
	"errors"
	"fmt"
	"strings"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"

	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	"github.com/0x0BSoD/torrBotGo/internal/transmission"
)

func handleMessage(update tgbotapi.Update, tClient *telegram.Client, trClient *transmission.Client) {
	tClient.SetChatID(update.Message.Chat.ID)

	if update.Message.Document != nil {
		_url, err := tClient.BotAPI.GetFileDirectURL(update.Message.Document.FileID)
		if err != nil {
			tClient.SendError(fmt.Sprintf("get file URL failed, %v", err))
			return
		}

		title, imgPath, err := trClient.AddByFileDialog(_url)
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
		return
	}

	if strings.HasPrefix(update.Message.Text, "magnet:") {
		text, err := trClient.AddByMagnetDialog(update.Message.Text)
		if err != nil {
			tClient.SendError(fmt.Sprintf("add torrent by magent link failed, %v", err))
			return
		}

		catList := extractKeys(trClient.Categories)
		kbdAdd := telegram.TorrentAddKbd(false, catList)

		if err := tClient.SendMessage(text, kbdAdd); err != nil {
			tClient.SendError(fmt.Sprintf("send dialog failed, %v", err))
			return
		}
		return
	}

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
