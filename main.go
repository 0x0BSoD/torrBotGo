package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
)

// GlobalContext - struct for keeping needed stuff, i dragged it through almost all functions
type GlobalContext struct {
	Bot          *tgbotapi.BotAPI
	TrAPI        *transmission.Client
	Mutex        sync.Mutex
	Debug        bool
	Categories   map[string]string
	TorrentCache torrents
	imgDir       string
	chatID       int64
}

var path string
var ctx GlobalContext

func init() {
	flag.StringVar(&path, "config", "./config.json", "path to config file, JSON")
}

func main() {
	flag.Parse()

	cfg := marshalConf(path)

	ctx.Debug = cfg.Debug
	b, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Fatalf("can't start session with telegram: %s", err)
	}
	b.Debug = cfg.Debug
	ctx.Bot = b
	ctx.Categories = cfg.Categories
	ctx.imgDir = cfg.ImgDir

	conf := transmission.Config{
		Address:  cfg.Transmission.URI,
		User:     cfg.Transmission.User,
		Password: cfg.Transmission.Password,
	}

	if ctx.Debug {
		_ = glh.PrettyPrint(cfg)
	}

	fmt.Print("Connecting to transmission API ")
	t, err := transmission.New(conf)
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

	if (transmission.SetSessionArgs{}) != cfg.CustomTransmissionArgs {
		fmt.Print("Setting custom transmission parameters ")
		err := t.Session.Set(cfg.CustomTransmissionArgs)
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
