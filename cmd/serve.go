package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"

	"github.com/0x0BSoD/torrBotGo/config"
	"github.com/0x0BSoD/torrBotGo/internal/events"
	"github.com/0x0BSoD/torrBotGo/internal/transmission"
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
		fmt.Println(fmt.Errorf("unable to create a config: %v", err))
		os.Exit(1)
	}

	config.EventBus = events.New(100)
	config.Logger = logger.New(zapcore.DebugLevel)

	config.Logger.Info("creating Telegram API client")
	b, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		config.Logger.Sugar().Errorf("can't create Telegram API client: %w", err)
		os.Exit(1)
	}
	config.Telegram.Client = b

	config.Logger.Info("connecting to transmission API")
	trCfg := transmission.Config{
		URI:      config.Transmission.Config.URI,
		User:     config.Transmission.Config.User,
		Password: config.Transmission.Config.Password,
		Custom:   config.Transmission.Custom,
		Logger:   config.Logger,
		EventBus: config.EventBus,
	}
	trClient, err := transmission.New(&trCfg)
	if err != nil {
		config.Logger.Sugar().Errorf("can't create Transmission API client: %w", err)
		os.Exit(1)
	}
	defer func() {
		config.Logger.Info("closing transmission session")
		config.Transmission.Client.API.HTTPClient.CloseIdleConnections()
	}()
	config.Transmission.Client = trClient

	config.Logger.Info("starting Event Bus")
	busCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go config.EventBus.Run(busCtx)

	fmt.Println(config)
}
