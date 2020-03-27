package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
)

type GlobalContext struct {
	Bot          *tgbotapi.BotAPI
	TrApi        *transmission.Client
	Mutex        sync.Mutex
	Debug        bool
	Categories   map[string]string
	TorrentCache Torrents
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
		log.Panic(err)
	}
	b.Debug = cfg.Debug
	ctx.Bot = b
	ctx.Categories = cfg.Categories

	conf := transmission.Config{
		Address:  cfg.Transmission.Uri,
		User:     cfg.Transmission.User,
		Password: cfg.Transmission.Password,
	}

	fmt.Print("Connecting to transmission API ")
	t, err := transmission.New(conf)
	if err != nil {
		fmt.Println("❌")
		log.Panic(err)
	}
	fmt.Println("✔️")
	defer func() {
		fmt.Print("Closing transmission session ")
		err := t.Session.Close()
		if err != nil {
			fmt.Println("❌")
			log.Panic(err)
		}
		fmt.Println("✔️")
	}()

	if cfg.DefaultDownloadDir != "" {
		err := t.Session.Set(transmission.SetSessionArgs{DownloadDir: cfg.DefaultDownloadDir})
		if err != nil {
			log.Panic(err)
		}
	}

	fmt.Print("Updating transmission session info ")
	err = t.Session.Update()
	if err != nil {
		fmt.Println("❌")
		log.Panic(err)
	}
	fmt.Println("✔️")

	fmt.Print("Setting torrents cache ")
	tMap, err := t.GetTorrentMap()
	if err != nil {
		fmt.Println("❌")
		log.Panic(err)
	}
	ctx.TorrentCache = InitCache(tMap)
	fmt.Println("✔️")

	ctx.TrApi = t

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	fmt.Print("Getting updates channel ")
	updates, err := ctx.Bot.GetUpdatesChan(u)
	if err != nil {
		fmt.Println("❌")
		log.Panic(err)
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
