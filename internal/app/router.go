// Package app provides the core application logic for torrBotGo.
// It handles Telegram message routing, command processing, and user interactions.
package app

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"

	"github.com/0x0BSoD/torrBotGo/config"
	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	"github.com/0x0BSoD/torrBotGo/internal/transmission"
)

type handler struct {
	tClient      *telegram.Client
	trClient     *transmission.Client
	autoCategory bool
}

// StartUpdateParser - loop for read updates from Telegram
func StartUpdateParser(ctx context.Context, cfg *config.Config, timeout time.Duration) error {
	// TODO: Implement persistent offset storage to avoid missing updates after restart.
	// Currently using offset 0 which retrieves all updates from the last 24 hours.
	// Should store the last processed update ID and resume from that point.
	u := tgbotapi.NewUpdate(0)
	u.Timeout = int(timeout)

	h := handler{
		tClient:      cfg.Telegram.Client,
		trClient:     cfg.Transmission.Client,
		autoCategory: cfg.App.AutoCategories,
	}

	updates, err := cfg.Telegram.Client.BotAPI.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("getting tg updates failed: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case upd := <-updates:
			switch {
			case upd.Message == nil:
				cfg.Logger.Sugar().Debugf("got inline message: %s", upd.CallbackQuery.Data)
				h.handleInline(upd)
			case upd.Message.IsCommand():
				cfg.Logger.Sugar().Debugf("got command message: %s", upd.Message.Command())
				h.handleCommand(upd)
			default:
				cfg.Logger.Sugar().Debugf("got plain message: %s", upd.Message.Text)
				h.handleMessage(upd)
			}
		}
	}
}
