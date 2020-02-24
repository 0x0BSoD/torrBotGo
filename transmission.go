package main

import (
	"bytes"
	"encoding/json"
	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
	"log"
	"text/template"
)

// ========================
// STATUS
//=========================

type status struct {
	Active     int
	Paused     int
	UploadS    string
	DownloadS  string
	Downloaded string
	Uploaded   string
}

func sendStatus() string {
	stats, err := ctx.TrApi.Session.Stats()
	if err != nil {
		log.Panic(err)
	}

	t, err := template.ParseFiles("templates/status.gotmpl")
	if err != nil {
		log.Panic(err)
	}

	if ctx.Debug {
		_ = glh.PrettyPrint(stats)
	}

	var dRes bytes.Buffer
	t.Execute(&dRes, status{
		Active:     stats.ActiveTorrentCount,
		Paused:     stats.PausedTorrentCount,
		UploadS:    glh.ConvertBytes(float64(stats.UploadSpeed), glh.Speed),
		DownloadS:  glh.ConvertBytes(float64(stats.DownloadSpeed), glh.Speed),
		Uploaded:   glh.ConvertBytes(float64(stats.CurrentStats.UploadedBytes), glh.Size),
		Downloaded: glh.ConvertBytes(float64(stats.CurrentStats.DownloadedBytes), glh.Size),
	})

	return dRes.String()
}

type sessConfig struct {
	DownloadDir   string
	StartAdded    bool
	SpeedLimitD   string
	SpeedLimitDEn bool
	SpeedLimitU   string
	SpeedLimitUEn bool
}

func sendConfig() string {
	ctx.TrApi.Session.Update()
	sc := ctx.TrApi.Session

	t, err := template.ParseFiles("templates/config.gotmpl")
	if err != nil {
		log.Panic(err)
	}

	if ctx.Debug {
		_ = glh.PrettyPrint(sc)
	}

	var dRes bytes.Buffer
	t.Execute(&dRes, sessConfig{
		DownloadDir:   sc.DownloadDir,
		StartAdded:    sc.StartAddedTorrents,
		SpeedLimitD:   glh.ConvertBytes(float64(sc.SpeedLimitDown), glh.Speed),
		SpeedLimitDEn: sc.SpeedLimitDownEnabled,
		SpeedLimitU:   glh.ConvertBytes(float64(sc.SpeedLimitUp), glh.Speed),
		SpeedLimitUEn: sc.SpeedLimitUpEnabled,
	})

	return dRes.String()
}

func sendJsonConfig() string {
	ctx.TrApi.Session.Update()
	sc := ctx.TrApi.Session

	if ctx.Debug {
		_ = glh.PrettyPrint(sc)
	}

	b, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		log.Panic(err)
	}

	return "```" + string(b) + "```"
}

type Torrent struct {
	Active      bool
	Name        string
	Status      string
	Icon        string
	Error       bool
	ErrorString string
	Size        string
	Comment     string
	Hash        string
	Dspeed      string
	Uspeed      string
}

func parseStatus(s int) (string, string) {
	var icon string
	var status string

	switch s {
	case 0:
		icon = "⏹️️"
		status = "Stopped"
	case 1:
		icon = "▶️️"
		status = "Queued to check files"
	case 2:
		icon = "▶️"
		status = "Checking files"
	case 3:
		icon = "▶️️"
		status = "Queued to download"
	case 4:
		icon = "▶️"
		status = "Downloading"
	case 5:
		icon = "▶️️"
		status = "'Queued to seed"
	default:
		icon = "▶️️"
		status = "Seeding"
	}

	return icon, status
}

type showFilter int

const (
	All showFilter = iota
	Active
	NotActive
)

func sendTorrent(id int64, torr *transmission.Torrent) {
	t, err := template.ParseFiles("templates/torrentList.gotmpl")
	if err != nil {
		log.Panic(err)
	}
	var dRes bytes.Buffer

	icon, status := parseStatus(torr.Status)

	t.Execute(&dRes, Torrent{
		Name:        torr.Name,
		Status:      status,
		Icon:        icon,
		ErrorString: torr.ErrorString,
		Comment:     torr.Comment,
		Hash:        torr.HashString,
	})

	msg := tgbotapi.NewMessage(id, dRes.String())
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = torrentKbd(torr.HashString)
	if msg.Text != "" {
		if _, err := ctx.Bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func sendTorrentList(id int64, sf showFilter) {
	torrents, err := ctx.TrApi.GetTorrents()
	if err != nil {
		log.Panic(err)
	}
	for _, i := range torrents {
		switch sf {
		case All:
			sendTorrent(id, i)
		case Active:
			if i.Status != 0 && i.ErrorString == "" {
				sendTorrent(id, i)
			}
		case NotActive:
			if i.Status == 0 {
				sendTorrent(id, i)
			}
		}
	}
}

func sendTorrentDetails(hash string) string {
	tMap, err := ctx.TrApi.GetTorrentMap()
	if err != nil {
		log.Panic(err)
	}

	var active bool
	var error bool
	if tMap[hash].Status != 0 {
		active = true
	}
	if tMap[hash].ErrorString != "" {
		error = true
	}
	icon, status := parseStatus(tMap[hash].Status)

	t, err := template.ParseFiles("templates/torrent.gotmpl")
	if err != nil {
		log.Panic(err)
	}

	if error {
		icon = "🔥️"
	}

	var dRes bytes.Buffer
	t.Execute(&dRes, Torrent{
		Active:      active,
		Error:       error,
		Name:        tMap[hash].Name,
		Status:      status,
		Icon:        icon,
		ErrorString: tMap[hash].ErrorString,
		Size:        glh.ConvertBytes(float64(tMap[hash].TotalSize), glh.Size),
		Dspeed:      glh.ConvertBytes(float64(tMap[hash].RateDownload), glh.Speed),
		Uspeed:      glh.ConvertBytes(float64(tMap[hash].RateUpload), glh.Speed),
	})

	return dRes.String()
}
