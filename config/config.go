package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/0x0BSoD/torrBotGo/internal/cache"
	"github.com/0x0BSoD/torrBotGo/internal/events"
	intTransmission "github.com/0x0BSoD/torrBotGo/internal/transmission"

	tgBotAPI "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Logger       *zap.Logger
	TorrentCache *cache.Torrents
	EventBus     *events.Bus
	App          struct {
		Debug      bool   `yaml:"debug"`
		ErrorMedia string `yaml:"error_media"`
		Dirs       struct {
			Images     string            `yaml:"images"`
			Working    string            `yaml:"working"`
			Download   string            `yaml:"default_download_dir"`
			Categories map[string]string `yaml:"categories"`
		} `yaml:"dirs"`
	} `yaml:"app"`
	Telegram struct {
		Client *tgBotAPI.BotAPI
		Token  string `yaml:"token"`
	} `yaml:"telegram"`
	Transmission struct {
		Config trConfig                    `yaml:"config"`
		Custom transmission.SetSessionArgs `yaml:"custom"`
		Client *intTransmission.Client
	} `yaml:"transmission"`
}

type trConfig struct {
	URI      string `yaml:"uri"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func New(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("config file not found, %w", err)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return Config{}, fmt.Errorf("can't read file, %w", err)
	}

	var result Config
	if err = yaml.Unmarshal(b, &result); err != nil {
		return Config{}, fmt.Errorf("config file marshal, %w", err)
	}

	if _, err := os.Stat(result.App.Dirs.Images); os.IsNotExist(err) {
		err = os.MkdirAll(result.App.Dirs.Images, 0o755)
		if err != nil {
			return Config{}, fmt.Errorf("can't create Images directory, %w", err)
		}
	}

	if _, err := os.Stat(result.App.Dirs.Download); os.IsNotExist(err) {
		err = os.MkdirAll(result.App.Dirs.Download, 0o755)
		if err != nil {
			return Config{}, fmt.Errorf("can't create Download directory, %w", err)
		}
	}

	for _, d := range result.App.Dirs.Categories {
		if _, err := os.Stat(result.App.Dirs.Download + d); os.IsNotExist(err) {
			err = os.MkdirAll(result.App.Dirs.Download+d, 0o755)
			if err != nil {
				return Config{}, fmt.Errorf("can't create directory, %w", err)
			}
		}
	}

	if !strings.HasSuffix(result.App.Dirs.Images, "/") {
		result.App.Dirs.Images += "/"
	}

	if _, err := os.Stat(result.App.Dirs.Working); os.IsNotExist(err) {
		return Config{}, fmt.Errorf("wrong Working directory, %w", err)
	}

	return result, nil
}

func configClients(config *Config) error {
	fmt.Println("FooBar")

	return nil
}
