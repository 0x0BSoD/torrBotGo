package telegram

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
)

// StartUpdateParser - loop for read updates from Telegram
func (c *Client) StartUpdateParser(ctx context.Context, interval time.Duration) error {
	// TODO: Store offset
	u := tgbotapi.NewUpdate(0)
	u.Timeout = int(interval)

	updates, err := c.BotAPI.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("getting tg updates failed: %s", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case upd := <-updates:
			if upd.Message == nil {
				c.handleInline(upd)
				return nil
			}

			if upd.Message.IsCommand() {
				c.handleCommand(upd)
				return nil
			}

			c.handleMessage(upd)
		}
	}
}
