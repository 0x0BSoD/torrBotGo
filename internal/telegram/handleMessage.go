package telegram

import (
	"strconv"
	"strings"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
)

func (c *Client) handleMessage(upd tgbotapi.Update) {
	var err error

	c.chatID = upd.Message.Chat.ID

	if torrentID, err := strconv.ParseInt(upd.Message.Text, 10, 64); err == nil {
		err = sendTorrentDetailsByID(torrentID)
		if err != nil {
			c.sendError(err.Error())
		}
		return
	}

	if upd.Message.Document != nil {
		err = addTorrentFileQuestion(upd.Message.Document.FileID, upd.Message.MessageID)
	} else if strings.HasPrefix(upd.Message.Text, "magnet:") {
		err = addTorrentMagnetQuestion(upd.Message.Text, upd.Message.MessageID)
	} else if strings.HasPrefix(upd.Message.Text, "t:") {
		err = searchTorrent(upd.Message.Text)
	} else {
		switch upd.Message.Text {
		case "All torrents":
			err = sendTorrentList(all)
		case "Active torrents":
			err = sendTorrentList(active)
		case "Not Active torrents":
			err = sendTorrentList(notActive)
		default:
			c.sendError("I don't know that command. handleMessage")
		}
	}

	if err != nil {
		c.sendError(err.Error())
	}
}
