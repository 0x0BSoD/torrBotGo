package transmission

import (
	"context"

	"github.com/0x0BSoD/torrBotGo/internal/cache"
	"github.com/0x0BSoD/torrBotGo/internal/events"
	"github.com/0x0BSoD/transmission"
	"go.uber.org/zap"
)

type Client struct {
	API      *transmission.Client
	logger   *zap.Logger
	eventBus *events.Bus
	cache    *cache.Torrents
	Storage  struct {
		Torrent    *transmission.Torrent
		tFile      []byte
		magentLink string
		messageID  int
	}
}

type Config struct {
	URI      string
	User     string
	Password string
	EventBus *events.Bus
	Logger   *zap.Logger
	Custom   transmission.SetSessionArgs
}

func New(cfg *Config) (*Client, error) {
	conf := transmission.Config{
		Address:  cfg.URI,
		User:     cfg.User,
		Password: cfg.Password,
	}

	t, err := transmission.New(conf)
	t.Context = context.TODO()
	if err != nil {
		return nil, err
	}

	if (transmission.SetSessionArgs{}) != cfg.Custom {
		cfg.Logger.Info("setting custom transmission parameters")
		err := t.Session.Set(cfg.Custom)
		if err != nil {
			cfg.Logger.Sugar().Errorf("getting tg updates failed: %w", err)
			return nil, err
		}
	}

	cfg.Logger.Info("updating transmission session info ")
	err = t.Session.Update()
	if err != nil {
		return nil, err
	}

	tMap, err := t.GetTorrentMap()
	if err != nil {
		return nil, err
	}
	cfg.Logger.Info("setting torrents cache")

	var result Client
	result.API = t
	result.logger = cfg.Logger
	result.eventBus = cfg.EventBus
	result.cache = cache.New(tMap)

	return &result, nil
}
