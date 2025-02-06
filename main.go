package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
)

// GlobalContext - struct for keeping needed stuff, I dragged it through almost all functions
type GlobalContext struct {
	Bot          *tgbotapi.BotAPI
	TrAPI        *transmission.Client
	Mutex        sync.Mutex
	Debug        bool
	Categories   map[string]string
	TorrentCache torrents
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
	// Config part =======
	flag.Parse()

	cfg := marshalConf(path)

	ctx.Debug = cfg.App.Debug

	b, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Fatalf("can't start session with telegram: %s", err)
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
	fmt.Print("Connecting to transmission API ")
	t, err := transmission.New(conf)
	t.Context = context.TODO()
	if err != nil {
		fmt.Println("❌")
		log.Fatalf(">> can't create transmission session: %s", err)
	}
	defer func() {
		fmt.Print("Closing transmission session ")
		err := t.Session.Close()
		if err != nil {
			fmt.Println("❌")
			log.Panic(err)
		}
		fmt.Println("✔️")
	}()

	if (transmission.SetSessionArgs{}) != cfg.Transmission.Custom {
		fmt.Print("Setting custom transmission parameters ")
		err := t.Session.Set(cfg.Transmission.Custom)
		if err != nil {
			fmt.Println("❌")
			log.Fatalf(">> starting transmission session failed: %s", err)
		}
	}
	fmt.Println("✔️")

	fmt.Print("Updating transmission session info ")
	err = t.Session.Update()
	if err != nil {
		fmt.Println("❌")
		log.Fatalf(">> updating transmission session info failed: %s", err)
	}
	fmt.Println("✔️")

	fmt.Print("Setting torrents cache ")
	tMap, err := t.GetTorrentMap()
	if err != nil {
		fmt.Println("❌")
		log.Fatalf(">> get torrent map failed: %s", err)
	}
	ctx.TorrentCache = initCache(tMap)
	fmt.Println("✔️")

	ctx.TrAPI = t

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	fmt.Print("Getting updates channel ")
	updates, err := ctx.Bot.GetUpdatesChan(u)
	if err != nil {
		fmt.Println("❌")
		log.Fatalf(">> getting tg updates failed: %s", err)
	}
	fmt.Println("✔️")

	fmt.Println("Bot started ✔️")

	go func() {
		for {
			updateCache()
			time.Sleep(1 * time.Minute)
		}
	}()

	for update := range updates {
		parseUpdate(update)
	}
}
