// Package ctx - GlobalContext struct for keeping needed stuff, I dragged it through almost all functions
package ctx

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
	"go.uber.org/zap"

	"github.com/0x0BSoD/torrBotGo/internal/cache"
)

type Telegram struct {
	Client     *tgbotapi.BotAPI
	Categories map[string]string
	ImgDir     string
	ErrMedia   string
	ChatID     int64
}

type GlobalContext struct {
	Telegram     Telegram
	TrAPI        *transmission.Client
	Debug        bool
	TorrentCache cache.Torrents
	Cwd          string
	Logger       *zap.Logger
}
