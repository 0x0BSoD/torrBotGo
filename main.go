package main

import (
	"flag"
	"log"
	"sync"

	"github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
)

type GlobalContext struct {
	Bot   *tgbotapi.BotAPI
	TrApi *transmission.Client
	Mutex sync.Mutex
	Debug bool
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
	ctx.Bot = b
	conf := transmission.Config{
		Address:  cfg.Transmission.Uri,
		User:     cfg.Transmission.User,
		Password: cfg.Transmission.Password,
	}
	t, err := transmission.New(conf)
	if err != nil {
		log.Panic(err)
	}
	ctx.TrApi = t

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := ctx.Bot.GetUpdatesChan(u)

	for update := range updates {
		parseUpdate(update)
	}
}
