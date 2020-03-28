package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type config struct {
	Token              string            `json:"token"`
	Debug              bool              `json:"debug"`
	Transmission       trConfig          `json:"transmission"`
	DefaultDownloadDir string            `json:"default_download_dir"`
	Categories         map[string]string `json:"categories"`
	ImgDir             string            `json:"img_dir"`
}

type trConfig struct {
	Uri      string `json:"uri"`
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
	err = json.Unmarshal(b, &result)
	if err != nil {
		fmt.Printf("can't parse file '%s', %s", path, err.Error())
		os.Exit(-1)
	}

	if _, err := os.Stat(result.ImgDir); os.IsNotExist(err) {
		err = os.MkdirAll(result.ImgDir, 0755)
		if err != nil {
			fmt.Printf("can't create ImgDir directory '%s', %s", result.ImgDir, err.Error())
			os.Exit(-1)
		}
	}

	if !strings.HasSuffix(result.ImgDir, "/") {
		result.ImgDir += "/"
	}

	return result
}
