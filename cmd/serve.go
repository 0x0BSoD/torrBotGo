// Package cmd provides the command-line interface for torrBotGo.
// It uses Cobra CLI framework to define and handle commands.
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"

	"github.com/0x0BSoD/torrBotGo/config"
	"github.com/0x0BSoD/torrBotGo/internal/app"
	"github.com/0x0BSoD/torrBotGo/internal/events"
	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	intTransmission "github.com/0x0BSoD/torrBotGo/internal/transmission"
	"github.com/0x0BSoD/torrBotGo/pkg/logger"
)

var configPath string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start Bot",
	Long:  `Start Bot`,
	Run:   serve,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&configPath, "config", "c", "./config.yaml", "Path to YAML config")
}

func serve(cmd *cobra.Command, args []string) {
	config, err := config.New(configPath)
	if err != nil {
		fmt.Println(fmt.Errorf("unable to create a config: %w", err))
		os.Exit(1)
	}

	config.EventBus = events.New(100)
	config.Logger = logger.New(zapcore.DebugLevel)

	config.Logger.Info("creating Telegram API client")
	tgClient, err := telegram.New(config.Telegram.Token, config.App.Dirs.Images, config.App.ErrorMedia, config.Logger)
	if err != nil {
		config.Logger.Sugar().Errorf("can't create Telegram API client: %w", err)
		os.Exit(1)
	}
	config.Telegram.Client = tgClient
	tgClient.SetChatID(config.Telegram.ChatID)

	config.Logger.Info("connecting to transmission API")
	trCfg := intTransmission.Config{
		URI:        config.Transmission.Config.URI,
		User:       config.Transmission.Config.User,
		Password:   config.Transmission.Config.Password,
		Custom:     config.Transmission.Custom,
		Logger:     config.Logger,
		EventBus:   config.EventBus,
		Categories: config.App.Dirs.Categories,
		MediaPath:  config.App.Dirs.Images,
	}
	trClient, err := intTransmission.New(&trCfg)
	if err != nil {
		config.Logger.Sugar().Errorf("can't create Transmission API client: %w", err)
		os.Exit(1)
	}
	defer func() {
		config.Logger.Info("closing transmission session")
		config.Transmission.Client.API.HTTPClient.CloseIdleConnections()
	}()
	config.Transmission.Client = trClient

	config.Logger.Info("starting Cache updater")
	cacheCtx, cacheCancel := context.WithCancel(context.Background())
	defer cacheCancel()
	go trClient.StartCacheUpdater(cacheCtx, intTransmission.CacheUpdateInterval)

	config.Logger.Info("starting Event Bus")
	busCtx, busCancel := context.WithCancel(context.Background())
	defer busCancel()
	go config.EventBus.Run(busCtx)
	config.EventBus.Subscribe(events.EventTorrentDownloadDone, func(ev events.Event) {
		_ = config.Telegram.Client.SendMessage(ev.Text, nil)
	})

	config.Logger.Info("starting TG update parser")
	prsrCtx, prsrCancel := context.WithCancel(context.Background())
	defer prsrCancel()
	app.StartUpdateParser(prsrCtx, &config, intTransmission.UpdateParserTimeout)
}
