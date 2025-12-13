package transmission

import (
	"context"
	"fmt"
	"time"

	"github.com/0x0BSoD/torrBotGo/internal/events"
	"github.com/0x0BSoD/transmission"
)

const doneEpsilon = 0.9999

// StartCacheUpdater - watcher that compare current torrent state and stored in memory
func (c *Client) StartCacheUpdater(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	c.logger.Debug("initial cache update")
	if err := c.updateCache(ctx); err != nil {
		c.logger.Sugar().Errorf("StartCacheUpdater: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := c.updateCache(ctx); err != nil {
				c.logger.Sugar().Errorf("StartCacheUpdater: %v", err)
			}
		}
	}
}

func (c *Client) updateCache(ctx context.Context) error {
	tMap, err := c.API.GetTorrentMap()
	if err != nil {
		return fmt.Errorf("updateCache: fetch: %w", err)
	}

	changed := c.cache.Update(tMap)
	if len(changed) == 0 {
		return nil
	}

	var msgs []string
	for _, t := range changed {
		if t == nil {
			continue
		}

		if t.ErrorString != "" {
			msgs = append(msgs,
				fmt.Sprintf("Failed\n%s\nError:\n%s", t.Name, t.ErrorString))
			continue
		}

		if t.PercentDone >= doneEpsilon && t.Status == transmission.StatusSeeding {
			msgs = append(msgs, fmt.Sprintf("Downloaded\n%s", t.Name))
		}
	}

	for _, m := range msgs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		c.eventBus.Publish(events.Event{
			Type: events.EventTorrentDownloadDone,
			Text: m,
		})
	}

	return nil
}
