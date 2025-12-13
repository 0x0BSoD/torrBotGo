package app

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"

	"github.com/0x0BSoD/torrBotGo/config"
)

// StartUpdateParser - loop for read updates from Telegram
func StartUpdateParser(ctx context.Context, cfg *config.Config, timeout time.Duration) error {
	// TODO: Store offset
	u := tgbotapi.NewUpdate(0)
	u.Timeout = int(timeout)

	updates, err := cfg.Telegram.Client.BotAPI.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("getting tg updates failed: %s", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case upd := <-updates:
			switch {
			case upd.Message == nil:
				cfg.Logger.Sugar().Debugf("got inline message: %s", upd.CallbackQuery.Data)
				handleInline(upd, cfg.Telegram.Client, cfg.Transmission.Client)
			case upd.Message.IsCommand():
				cfg.Logger.Sugar().Debugf("got command message: %s", upd.Message.Command())
				handleCommand(upd, cfg.Telegram.Client, cfg.Transmission.Client)
			default:
				cfg.Logger.Sugar().Debugf("got plain message: %s", upd.Message.Text)
				handleMessage(upd, cfg.Telegram.Client, cfg.Transmission.Client)
			}
		}
	}
}
