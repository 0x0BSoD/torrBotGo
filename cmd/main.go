package main

import (
	"context"
	"flag"
	"time"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
	"go.uber.org/zap/zapcore"

	"github.com/0x0BSoD/torrBotGo/internal/cache"
	_ctx "github.com/0x0BSoD/torrBotGo/internal/ctx"
	"github.com/0x0BSoD/torrBotGo/pkg/logger"
)

var (
	path string
	ctx  _ctx.GlobalContext
)

func init() {
	flag.StringVar(&path, "config", "./config.yaml", "path to config file, YAML")
	flag.StringVar(&path, "c", "./config.yaml", "path to config file, YAML")
}

func main() {
	// Config part =======
	flag.Parse()

	log := logger.New(zapcore.DebugLevel)

	cfg := marshalConf(path)

	ctx.Debug = cfg.App.Debug

	b, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Sugar().Errorf("can't start session with telegram: %s", err)
	}
	b.Debug = cfg.App.Debug

	ctx.Bot = b
	ctx.Categories = cfg.App.Dirs.Categories
	ctx.Cwd = cfg.App.WorkingDir
	ctx.ImgDir = cfg.App.ImgDir
	ctx.ErrMedia = cfg.App.ErrorMedia

	conf := transmission.Config{
		Address:  cfg.Transmission.Config.URI,
		User:     cfg.Transmission.Config.User,
		Password: cfg.Transmission.Config.Password,
	}

	if ctx.Debug {
		_ = glh.PrettyPrint(cfg)
	}

	// App run part ====
	log.Info("connecting to transmission API")
	t, err := transmission.New(conf)
	t.Context = context.TODO()
	if err != nil {
		log.Sugar().Errorf("can't create transmission session: %s", err)
	}
	defer func() {
		log.Info("closing transmission session")
		err := t.Session.Close()
		if err != nil {
			log.Sugar().Error(err)
		}
	}()

	if (transmission.SetSessionArgs{}) != cfg.Transmission.Custom {
		log.Info("setting custom transmission parameters")
		err := t.Session.Set(cfg.Transmission.Custom)
		if err != nil {
			log.Sugar().Errorf("starting transmission session failed: %s", err)
		}
	}

	log.Info("updating transmission session info ")
	err = t.Session.Update()
	if err != nil {
		log.Sugar().Errorf("updating transmission session info failed: %s", err)
	}

	log.Info("setting torrents cache")
	tMap, err := t.GetTorrentMap()
	if err != nil {
		log.Sugar().Errorf("get torrent map failed: %s", err)
	}
	ctx.TorrentCache = *cache.New(tMap)

	ctx.TrAPI = t

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	log.Info("getting updates channel")
	updates, err := ctx.Bot.GetUpdatesChan(u)
	if err != nil {
		log.Sugar().Errorf("getting tg updates failed: %s", err)
	}

	log.Debug("bot started")

	// TODO: ADD NEW WATCHER
	go func() {
		for {
			ctx.TorrentCache.Update()
			time.Sleep(1 * time.Minute)
		}
	}()

	for update := range updates {
		parseUpdate(update)
	}
}
