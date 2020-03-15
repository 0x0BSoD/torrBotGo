package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
)

type GlobalContext struct {
	Bot        *tgbotapi.BotAPI
	TrApi      *transmission.Client
	Mutex      sync.Mutex
	Debug      bool
	Categories map[string]string
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
	b.Debug = true
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
	defer t.Session.Close()

	fmt.Print("Updating transmission session info ")
	err = t.Session.Update()
	if err != nil {
		fmt.Println("❌")
		log.Panic(err)
	}
	fmt.Println("✔️")

	if cfg.DefaultDownloadDir != "" {
		err := t.Session.Set(transmission.SetSessionArgs{DownloadDir: cfg.DefaultDownloadDir})
		if err != nil {
			log.Panic(err)
		}
	}

	ctx.TrApi = t

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	fmt.Print("Getting updates chain ")
	updates, err := ctx.Bot.GetUpdatesChan(u)
	if err != nil {
		fmt.Println("❌")
		log.Panic(err)
	}
	fmt.Println("✔️")

	fmt.Println("Bot started ✔️")
	for update := range updates {
		parseUpdate(update)
	}
}
