package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/0x0BSoD/transmission"
	"gopkg.in/yaml.v2"
)

type config struct {
	App struct {
		Debug      bool   `yaml:"debug"`
		ImgDir     string `yaml:"img_dir"`
		WorkingDir string `yaml:"working_dir"`
		ErrorMedia string `yaml:"error_media"`
		Dirs       struct {
			DefaultDownloadDir string            `yaml:"default_download_dir"`
			Categories         map[string]string `yaml:"categories"`
		} `yaml:"dirs"`
	} `yaml:"app"`
	Telegram struct {
		Token string `yaml:"token"`
	} `yaml:"telegram"`
	Transmission struct {
		Config trConfig                    `yaml:"config"`
		Custom transmission.SetSessionArgs `yaml:"custom"`
	} `yaml:"transmission"`
}

type trConfig struct {
	URI      string `yaml:"uri"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func marshalConf(path string) (config, error) {
	f, err := os.Open(path)
	if err != nil {
		return config{}, fmt.Errorf("config file not found, %w", err)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return config{}, fmt.Errorf("can't read file, %w", err)
	}

	var result config
	if err = yaml.Unmarshal(b, &result); err != nil {
		return config{}, fmt.Errorf("config file marshal, %w", err)
	}

	if _, err := os.Stat(result.App.ImgDir); os.IsNotExist(err) {
		err = os.MkdirAll(result.App.ImgDir, 0o755)
		if err != nil {
			return config{}, fmt.Errorf("can't create ImgDir directory, %w", err)
		}
	}

	if _, err := os.Stat(result.App.Dirs.DefaultDownloadDir); os.IsNotExist(err) {
		err = os.MkdirAll(result.App.ImgDir, 0o755)
		if err != nil {
			return config{}, fmt.Errorf("can't create DefaultDownloadDir directory, %w", err)
		}
	}

	for _, d := range result.App.Dirs.Categories {
		if _, err := os.Stat(result.App.Dirs.DefaultDownloadDir + d); os.IsNotExist(err) {
			err = os.MkdirAll(result.App.Dirs.DefaultDownloadDir+d, 0o755)
			if err != nil {
				return config{}, fmt.Errorf("can't create directory, %w", err)
			}
		}
	}

	if !strings.HasSuffix(result.App.ImgDir, "/") {
		result.App.ImgDir += "/"
	}

	if _, err := os.Stat(result.App.WorkingDir); os.IsNotExist(err) {
		return config{}, fmt.Errorf("wrong WorkingDir directory, %w", err)
	}

	return result, nil
}
