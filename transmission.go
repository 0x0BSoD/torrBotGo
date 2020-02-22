package main

import (
	"bytes"
	"encoding/json"
	glh "github.com/0x0BSoD/goLittleHelpers"
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
