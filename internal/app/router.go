package app

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"

	"github.com/0x0BSoD/torrBotGo/config"
)

// StartUpdateParser - loop for read updates from Telegram
func StartUpdateParser(ctx context.Context, cfg *config.Config, interval time.Duration) error {
	// TODO: Store offset
	u := tgbotapi.NewUpdate(0)
	u.Timeout = int(interval)

	updates, err := cfg.Telegram.Client.BotAPI.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("getting tg updates failed: %s", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case upd := <-updates:
			if upd.Message.IsCommand() {
				handleCommand(upd, cfg.Telegram.Client, cfg.Transmission.Client, cfg.Logger)
			}
		}
	}
}
