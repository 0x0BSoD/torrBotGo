package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/torrBotGo/config"
	"github.com/0x0BSoD/torrBotGo/internal/events"
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

	fmt.Println(config)
}
