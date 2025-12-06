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

func (c *Client) SendMessage(chatID int64, text string, replyMarkup any) error {
	if chatID == 0 {
		return errors.New("chatID is empty")
	}

	if text == "" {
		return fmt.Errorf("message cannot be empty")
	}

	msg := tgbotapi.NewMessage(chatID, escapeAll(text))
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

func (c *Client) SendImagedMessage(chatID int64, text string, imgPath, replyMarkup any) error {
	msg := tgbotapi.NewPhotoUpload(chatID, imgPath)
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

func (c *Client) SendError(chatID int64, text string) {
	if chatID == 0 {
		c.logger.Sugar().Errorf("chatID empty, %s", text)
		return
	}

	msg := tgbotapi.NewPhotoUpload(chatID, c.errorMediaPath)
	msg.Caption = text

	if _, err := c.BotAPI.Send(msg); err != nil {
		c.logger.Sugar().Panicf("SendError failed, %s", err)
	}
}
