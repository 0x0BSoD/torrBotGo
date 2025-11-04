package telegram

import (
	"log"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
)

func (c *Client) handleCommand(upd tgbotapi.Update) {
	c.chatID = upd.Message.Chat.ID
	msg := tgbotapi.NewMessage(c.chatID, "")
	msg.ParseMode = "MarkdownV2"
	var err error
	var text string

	switch upd.Message.Command() {
	case "help", "start":
		text = "Telegram Bot as interface for transmission"
		msg.ReplyMarkup = mainKbd
	case "status":
		text, err = sendStatus()
	case "config":
		text, err = sendConfig()
		msg.ReplyMarkup = configKbd
	default:
		c.sendError("I don't know that command. handleCommand")
		return
	}

	msg.Text = text

	if err != nil {
		c.sendError(err.Error())
	} else {
		if msg.Text != "" {
			if _, err := c.BotAPI.Send(msg); err != nil {
				log.Panic(err)
			}
		}
	}
}
