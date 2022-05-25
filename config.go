package main

import (
	"fmt"
	"github.com/0x0BSoD/transmission"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
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
	URI      string `json:"uri"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func marshalConf(path string) config {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("config file '%s' not found, %s", path, err.Error())
		os.Exit(-1)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Printf("can't read file '%s', %s", path, err.Error())
		os.Exit(-1)
	}

	var result config
	if err = yaml.Unmarshal(b, &result); err != nil {
		fmt.Printf("can't parse file '%s', %s", path, err.Error())
		os.Exit(-1)
	}

	if _, err := os.Stat(result.App.ImgDir); os.IsNotExist(err) {
		fmt.Printf("❌ %s not exist \n", result.App.ImgDir)
		err = os.MkdirAll(result.App.ImgDir, 0755)
		if err != nil {
			fmt.Printf("can't create ImgDir directory '%s', %s", result.App.ImgDir, err.Error())
			os.Exit(-1)
		}
	}

	if _, err := os.Stat(result.App.Dirs.DefaultDownloadDir); os.IsNotExist(err) {
		fmt.Printf("❌ %s not exist \n", result.App.Dirs.DefaultDownloadDir)
		err = os.MkdirAll(result.App.ImgDir, 0755)
		if err != nil {
			fmt.Printf("can't create DefaultDownloadDir directory '%s', %s", result.App.ImgDir, err.Error())
			os.Exit(-1)
		}
	}

	for _, d := range result.App.Dirs.Categories {
		if _, err := os.Stat(result.App.Dirs.DefaultDownloadDir + d); os.IsNotExist(err) {
			fmt.Printf("❌ %s not exist \n", result.App.Dirs.DefaultDownloadDir+d)
			err = os.MkdirAll(result.App.Dirs.DefaultDownloadDir+d, 0755)
			if err != nil {
				fmt.Printf("can't create %s directory '%s', %s", d, result.App.ImgDir, err.Error())
				os.Exit(-1)
			}
		}
	}

	if !strings.HasSuffix(result.App.ImgDir, "/") {
		result.App.ImgDir += "/"
	}

	if _, err := os.Stat(result.App.WorkingDir); os.IsNotExist(err) {
		fmt.Printf("wrong WorkingDir directory '%s', %s", result.App.ImgDir, err.Error())
		os.Exit(-1)
	}

	return result
}
