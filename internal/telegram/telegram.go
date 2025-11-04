package telegram

import (
	"fmt"

	"go.uber.org/zap"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
)

type Client struct {
	BotAPI *tgbotapi.BotAPI
	chatID int64
	logger *zap.Logger
}

func (c *Client) Init(token string, log *zap.Logger) error {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("can't start session with telegram: %s", err)
	}

	c.BotAPI = b
	c.logger = log

	return nil
}
