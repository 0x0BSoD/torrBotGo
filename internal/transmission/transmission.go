package transmission

import (
	"context"
	"encoding/json"
	"fmt"

	glh "github.com/0x0BSoD/goLittleHelpers"
	"go.uber.org/zap"

	"github.com/0x0BSoD/torrBotGo/internal/cache"
	"github.com/0x0BSoD/torrBotGo/internal/events"
	"github.com/0x0BSoD/transmission"
)

type Client struct {
	API        *transmission.Client
	Categories map[string]string
	logger     *zap.Logger
	eventBus   *events.Bus
	cache      *cache.Torrents
	mediaPath  string
	Storage    struct {
		Torrent    *transmission.Torrent
		tFile      []byte
		magentLink string
		messageID  int
	}
}

type Config struct {
	URI        string
	User       string
	Password   string
	EventBus   *events.Bus
	Logger     *zap.Logger
	Categories map[string]string
	MediaPath  string
	Custom     transmission.SetSessionArgs
}

type Status struct {
	Active     int
	Paused     int
	UploadS    string
	DownloadS  string
	Downloaded string
	Uploaded   string
	FreeSpace  string
}

type SessionConfig struct {
	DownloadDir   string
	StartAdded    bool
	SpeedLimitD   string
	SpeedLimitDEn bool
	SpeedLimitU   string
	SpeedLimitUEn bool
	DownloadQEn   bool
	DownloadQSize int
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
	result.Categories = cfg.Categories
	result.mediaPath = cfg.MediaPath

	cfg.Logger.Info("updating transmission session info ")
	err = t.Session.Update()
	if err != nil {
		cfg.Logger.Sugar().Errorf("updating transmission session info failed: %s", err)
	}

	return &result, nil
}

func (c *Client) Status() (Status, error) {
	c.logger.Info("get transmission status")
	stats, err := c.API.Session.Stats()
	if err != nil {
		return Status{}, err
	}

	downloadDir := c.API.Session.DownloadDir
	freeSpaceData, err := c.API.FreeSpace(downloadDir)
	if err != nil {
		return Status{}, fmt.Errorf("error with %s: %s", downloadDir, err)
	}

	return Status{
		Active:     stats.ActiveTorrentCount,
		Paused:     stats.PausedTorrentCount,
		UploadS:    glh.ConvertBytes(float64(stats.UploadSpeed), glh.Speed),
		DownloadS:  glh.ConvertBytes(float64(stats.DownloadSpeed), glh.Speed),
		Uploaded:   glh.ConvertBytes(float64(stats.CurrentStats.UploadedBytes), glh.Size),
		Downloaded: glh.ConvertBytes(float64(stats.CurrentStats.DownloadedBytes), glh.Size),
		FreeSpace:  glh.ConvertBytes(float64(freeSpaceData), glh.Size),
	}, nil
}

func (c *Client) SessionConfig() (SessionConfig, error) {
	c.logger.Info("get transmission config")
	err := c.API.Session.Update()
	if err != nil {
		return SessionConfig{}, err
	}

	sc := c.API.Session

	return SessionConfig{
		DownloadDir:   sc.DownloadDir,
		StartAdded:    sc.StartAddedTorrents,
		SpeedLimitD:   glh.ConvertBytes(float64(sc.SpeedLimitDown), glh.Speed),
		SpeedLimitDEn: sc.SpeedLimitDownEnabled,
		SpeedLimitU:   glh.ConvertBytes(float64(sc.SpeedLimitUp), glh.Speed),
		SpeedLimitUEn: sc.SpeedLimitUpEnabled,
		DownloadQEn:   sc.DownloadQueueEnabled,
		DownloadQSize: sc.DownloadQueueSize,
	}, nil
}

func (c *Client) SessionJSONConfig() (string, error) {
	c.logger.Info("get transmission JSON config")
	err := c.API.Session.Update()
	if err != nil {
		return "", err
	}
	sc := c.API.Session

	b, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("`%s`", string(b)), nil
}
