package ctx

import (
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	tr "github.com/0x0BSoD/transmission"
	"go.uber.org/zap"

	"github.com/0x0BSoD/torrBotGo/internal/cache"
)

type telegram struct {
	Client     *tgbotapi.BotAPI
	Categories map[string]string
	ImgDir     string
	ErrMedia   string
	ChatID     int64
}

type transmission struct {
	Client *tr.Client
}

type GlobalContext struct {
	Telegram     telegram
	Transmission transmission
	Debug        bool
	TorrentCache cache.Torrents
	Cwd          string
	Logger       *zap.Logger
}
