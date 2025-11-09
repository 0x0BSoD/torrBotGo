package main

import (
	"context"
	"flag"
	"os"
	"time"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/torrBotGo/internal/cache"
	"github.com/0x0BSoD/torrBotGo/pkg/logger"
	"go.uber.org/zap/zapcore"
)

// GlobalContext - struct for keeping needed stuff, I dragged it through almost all functions
type GlobalContext struct {
	Bot          *tgbotapi.BotAPI
	Transmisson  *trClient
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

	if ctx.Debug {
		_ = glh.PrettyPrint(cfg)
	}

	log.Info("connecting to transmission API")
	ctx.Transmisson, err = trInit(&cfg, log)
	if err != nil {
		log.Sugar().Errorf("can't create Transmission API client: %w", err)
		os.Exit(1)
	}
	defer func() {
		log.Info("closing transmission session")
		err := ctx.Transmisson.Api.Session.Close()
		if err != nil {
			log.Panic(err.Error())
		}
	}()

	// App run part ====
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	log.Info("getting updates channel")
	updates, err := ctx.Bot.GetUpdatesChan(u)
	if err != nil {
		log.Sugar().Errorf("getting tg updates failed: %s", err)
		os.Exit(1)
	}

	log.Info("starting cahce updater")
	go ctx.Transmisson.startCacheUpdater(context.TODO(), 1*time.Minute, &ctx)

	log.Debug("bot started")
	for update := range updates {
		parseUpdate(update)
	}
}
