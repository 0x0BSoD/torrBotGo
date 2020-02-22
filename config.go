package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type config struct {
	Token        string   `json:"token"`
	Debug        bool     `json:"debug"`
	Transmission trConfig `json:"transmission"`
}

type trConfig struct {
	Uri      string `json:"uri"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func marshalConf(path string) config {
	f, err := os.Open(path)
	if err != nil {
		log.Panicf("config file '%s' not found, %s", path, err.Error())
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Panicf("can't read file '%s', %s", path, err.Error())
	}

	var result config
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Panicf("can't parse file '%s', %s", path, err.Error())
	}

	return result
}
