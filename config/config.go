package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/0x0BSoD/torrBotGo/internal/events"
	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	intTransmission "github.com/0x0BSoD/torrBotGo/internal/transmission"
	"github.com/0x0BSoD/transmission"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Logger   *zap.Logger
	EventBus *events.Bus

	App struct {
		ErrorMedia     string `yaml:"error_media"`
		AutoCategories bool   `yaml:"auto_categories"`
		Dirs           struct {
			Images     string `yaml:"images"`
			Working    string `yaml:"working"`
			Download   string `yaml:"download"`
			Categories map[string]struct {
				Path    string `yaml:"path"`
				Matcher string `yaml:"matcher"`
			} `yaml:"categories"`
		} `yaml:"dirs"`
	} `yaml:"app"`

	Telegram struct {
		Client *telegram.Client
		Token  string `yaml:"token"`
	} `yaml:"telegram"`

	Transmission struct {
		Config struct {
			URI      string `yaml:"uri"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
		} `yaml:"config"`
		Custom transmission.SetSessionArgs `yaml:"custom,omitempty"`
		Client *intTransmission.Client
	} `yaml:"transmission"`
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
		path := filepath.Join(result.App.Dirs.Download, d.Path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = os.MkdirAll(path, 0o755)
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
