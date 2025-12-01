package telegram

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"go.uber.org/zap"
)

type Client struct {
	BotAPI *tgbotapi.BotAPI
	chatID int64
	logger *zap.Logger
}

func New(token string, log *zap.Logger) (*Client, error) {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	var result Client
	result.BotAPI = b
	result.logger = log

	return &result, nil
}
