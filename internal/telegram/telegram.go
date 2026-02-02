// Package telegram provides Telegram Bot API integration for torrBotGo.
// It handles message sending, inline keyboards, and communication with Telegram servers.
package telegram

import (
	"errors"
	"fmt"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"go.uber.org/zap"
)

type Client struct {
	BotAPI         *tgbotapi.BotAPI
	errorMediaPath string
	mediaPath      string
	logger         *zap.Logger
	storage        struct {
		chatID int64
	}
}

func New(token, mediaPath, errorMediaPath string, log *zap.Logger) (*Client, error) {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	var result Client
	result.BotAPI = b
	result.logger = log
	result.errorMediaPath = errorMediaPath
	result.mediaPath = mediaPath

	return &result, nil
}

func (c *Client) SetChatID(chatID int64) {
	c.storage.chatID = chatID
}

func (c *Client) SendMessage(text string, replyMarkup any) error {
	if c.storage.chatID == 0 {
		return errors.New("chatID is empty")
	}

	if text == "" {
		return fmt.Errorf("message cannot be empty")
	}

	msg := tgbotapi.NewMessage(c.storage.chatID, escapeAll(text))
	msg.ParseMode = "MarkdownV2"

	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup
	}

	c.logger.Info("send message to telegram")
	if _, err := c.BotAPI.Send(msg); err != nil {
		return err
	}

	return nil
}

func (c *Client) SendImagedMessage(text string, imgPath, replyMarkup any) error {
	msg := tgbotapi.NewPhotoUpload(c.storage.chatID, imgPath)
	msg.ParseMode = "MarkdownV2"

	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup
	}

	msg.Caption = escapeAll(text)
	if _, err := c.BotAPI.Send(msg); err != nil {
		return err
	}

	return nil
}

func (c *Client) SendError(text string) {
	if c.storage.chatID == 0 {
		c.logger.Sugar().Errorf("chatID empty, %s", text)
		return
	}

	msg := tgbotapi.NewPhotoUpload(c.storage.chatID, c.errorMediaPath)
	msg.Caption = text

	if _, err := c.BotAPI.Send(msg); err != nil {
		c.logger.Sugar().Panicf("SendError failed, %s", err)
	}
}

func (c *Client) RemoveMessage(messageID int) error {
	msgRm := tgbotapi.NewDeleteMessage(c.storage.chatID, messageID)

	if _, err := c.BotAPI.Send(msgRm); err != nil {
		return err
	}

	return nil
}

func (c *Client) SendEditedMessage(messageID int, text string, replyMarkup any) error {
	if text == "" {
		return fmt.Errorf("message cannot be empty")
	}

	msg := tgbotapi.NewEditMessageText(c.storage.chatID, messageID, escapeAll(text))
	msg.ParseMode = "MarkdownV2"

	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup.(*tgbotapi.InlineKeyboardMarkup)
	}

	if _, err := c.BotAPI.Send(msg); err != nil {
		return err
	}

	return nil
}
