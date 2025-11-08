package main

import (
	"context"
	"flag"
	"os"
	"sync"
	"time"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/torrBotGo/internal/cache"
	"github.com/0x0BSoD/torrBotGo/pkg/logger"
	"github.com/0x0BSoD/transmission"
	"go.uber.org/zap/zapcore"
)

// GlobalContext - struct for keeping needed stuff, I dragged it through almost all functions
type GlobalContext struct {
	Bot          *tgbotapi.BotAPI
	TrAPI        *transmission.Client
	Mutex        sync.Mutex
	Debug        bool
	Categories   map[string]string
	TorrentCache *cache.Torrents
	imgDir       string
	errMedia     string
	chatID       int64
	wd           string
}

var (
	path string
	ctx  GlobalContext
)

func init() {
	flag.StringVar(&path, "config", "./config.yaml", "path to config file, YAML")
	flag.StringVar(&path, "c", "./config.yaml", "path to config file, YAML")
}

func main() {
	log := logger.New(zapcore.DebugLevel)

	// Config part =======
	flag.Parse()

	log.Info("configuration init")
	cfg, err := marshalConf(path)
	if err != nil {
		log.Sugar().Errorf("can't init configuration: %s", err)
		os.Exit(1)
	}

	ctx.Debug = cfg.App.Debug

	log.Info("creating Telegram API client")
	b, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Sugar().Errorf("can't create Telegram API client: %w", err)
		os.Exit(1)
	}
	b.Debug = cfg.App.Debug

	ctx.Bot = b
	ctx.Categories = cfg.App.Dirs.Categories
	ctx.wd = cfg.App.WorkingDir
	ctx.imgDir = cfg.App.ImgDir
	ctx.errMedia = cfg.App.ErrorMedia

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
		log.Sugar().Errorf("can't create transmission session: %w", err)
		os.Exit(1)
	}
	defer func() {
		err := t.Session.Close()
		if err != nil {
			log.Panic(err.Error())
		}
	}()

	if (transmission.SetSessionArgs{}) != cfg.Transmission.Custom {
		log.Info("setting custom transmission parameters")
		err := t.Session.Set(cfg.Transmission.Custom)
		if err != nil {
			log.Sugar().Errorf("getting tg updates failed: %w", err)
		}
	}

	log.Info("updating transmission session info ")
	err = t.Session.Update()
	if err != nil {
		log.Sugar().Errorf("updating transmission session info failed: %s", err)
		os.Exit(1)
	}

	log.Info("setting torrents cache")
	tMap, err := t.GetTorrentMap()
	if err != nil {
		log.Sugar().Errorf("getting tg updates failed: %s", err)
	}
	ctx.TorrentCache = cache.New(tMap)

	ctx.TrAPI = t

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	log.Info("getting updates channel")
	updates, err := ctx.Bot.GetUpdatesChan(u)
	if err != nil {
		log.Sugar().Errorf("getting tg updates failed: %s", err)
		os.Exit(1)
	}

	log.Info("starting cahce updater")
	go startCacheUpdater(context.TODO(), 1*time.Minute, &ctx)

	log.Debug("bot started")
	for update := range updates {
		parseUpdate(update)
	}
}
