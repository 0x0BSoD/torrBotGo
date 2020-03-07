package main

import (
	"bytes"
	"encoding/json"
	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
	"log"
	"sort"
	"text/template"
)

var TORRENT *transmission.Torrent

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

func sendStatus() (string, error) {
	stats, err := ctx.TrApi.Session.Stats()
	if err != nil {
		return "", err
	}

	t, err := template.ParseFiles("templates/status.gotmpl")
	if err != nil {
		return "", err
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

	return dRes.String(), nil
}

type sessConfig struct {
	DownloadDir   string
	StartAdded    bool
	SpeedLimitD   string
	SpeedLimitDEn bool
	SpeedLimitU   string
	SpeedLimitUEn bool
}

func sendConfig() (string, error) {
	ctx.TrApi.Session.Update()
	sc := ctx.TrApi.Session

	t, err := template.ParseFiles("templates/config.gotmpl")
	if err != nil {
		return "", err
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

	return dRes.String(), nil
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
	Active         bool
	Name           string
	Status         string
	Icon           string
	Error          bool
	ErrorString    string
	DownloadedSize string
	Size           string
	Comment        string
	Hash           string
	PosInQ         int
	Dspeed         string
	Uspeed         string
}

type showFilter int

const (
	All showFilter = iota
	Active
	NotActive
)

//======================================================================================================================
// GET
//======================================================================================================================

func sendTorrent(id int64, torr *transmission.Torrent) error {
	t, err := template.ParseFiles("templates/torrentList.gotmpl")
	if err != nil {
		return err
	}

	icon, status := parseStatus(torr.Status)
	var dRes bytes.Buffer
	err = t.Execute(&dRes, Torrent{
		Name:        torr.Name,
		Status:      status,
		Icon:        icon,
		ErrorString: torr.ErrorString,
		Comment:     torr.Comment,
		Hash:        torr.HashString,
		PosInQ:      torr.QueuePosition,
	})
	if err != nil {
		return err
	}

	replyMarkup := torrentKbd(torr.HashString, torr.Status)
	err = sendNewMessage(id, dRes.String(), &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func sendTorrentList(id int64, sf showFilter) error {
	torrents, err := ctx.TrApi.GetTorrents()
	if err != nil {
		return err
	}

	sort.Slice(torrents[:], func(i, j int) bool {
		return torrents[i].QueuePosition < torrents[i].QueuePosition
	})

	for _, i := range torrents {
		switch sf {
		case All:
			err := sendTorrent(id, i)
			if err != nil {
				return err
			}
		case Active:
			if i.Status != 0 && i.ErrorString == "" {
				err := sendTorrent(id, i)
				if err != nil {
					return err
				}
			}
		case NotActive:
			if i.Status == 0 {
				err := sendTorrent(id, i)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func getTorrentDetails(hash string) string {
	if TORRENT == nil || TORRENT.HashString != hash {
		tMap, err := ctx.TrApi.GetTorrentMap()
		if err != nil {
			log.Panic(err)
		}
		TORRENT = tMap[hash]
	}

	var active bool
	if TORRENT.Status != 0 {
		active = true
	}

	icon, status := parseStatus(TORRENT.Status)
	var _error bool
	if TORRENT.ErrorString != "" {
		_error = true
		icon = "🔥️"
	}

	t, err := template.ParseFiles("templates/torrent.gotmpl")
	if err != nil {
		log.Panic(err)
	}

	var dRes bytes.Buffer
	t.Execute(&dRes, Torrent{
		Active:         active,
		Error:          _error,
		Name:           TORRENT.Name,
		Status:         status,
		Icon:           icon,
		ErrorString:    TORRENT.ErrorString,
		Size:           glh.ConvertBytes(float64(TORRENT.TotalSize), glh.Size),
		DownloadedSize: glh.ConvertBytes(float64(TORRENT.LeftUntilDone), glh.Size),
		Dspeed:         glh.ConvertBytes(float64(TORRENT.RateDownload), glh.Speed),
		Uspeed:         glh.ConvertBytes(float64(TORRENT.RateUpload), glh.Speed),
	})

	return dRes.String()
}

type filesList struct {
	Name        string
	Size        string
	Downloading bool
}

func sendTorrentDetails(hash string, chatID int64, messageID int) error {
	t := getTorrentDetails(hash)
	replyMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err := sendEditedMessage(chatID, messageID, t, &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func sendTorrentFiles(hash string, ID int64) {
	tMap, err := ctx.TrApi.GetTorrentMap()
	if err != nil {
		log.Panic(err)
	}

	files := *tMap[hash].Files
	filesStats := *tMap[hash].FileStats

	for i := 0; i < len(files); i++ {

		msg := tgbotapi.NewMessage(ID, "")
		msg.ParseMode = "MarkdownV2"

		t, err := template.ParseFiles("templates/torrentFile.gotmpl")
		if err != nil {
			log.Panic(err)
		}

		var dRes bytes.Buffer
		t.Execute(&dRes, filesList{
			Name:        files[i].Name,
			Size:        glh.ConvertBytes(float64(files[i].Length), glh.Size),
			Downloading: filesStats[i].Wanted,
		})

		msg.Text = dRes.String()

		if msg.Text != "" {
			if _, err := ctx.Bot.Send(msg); err != nil {
				log.Panic(err)
			}
		}
	}

}

//======================================================================================================================
// ACTIONS WITH TORRENT
//======================================================================================================================

func stopTorrent(hash string, cID int64, mID int) error {
	err := TORRENT.Stop()
	if err != nil {
		return err
	}
	TORRENT = nil

	msgTxt := getTorrentDetails(hash)
	newMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendEditedMessage(cID, mID, msgTxt, &newMarkup)
	if err != nil {
		return err
	}

	return nil
}

func startTorrent(hash string, cID int64, mID int) error {
	err := TORRENT.Start()
	if err != nil {
		return err
	}
	TORRENT = nil

	msgTxt := getTorrentDetails(hash)
	newMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendEditedMessage(cID, mID, msgTxt, &newMarkup)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Add this in API
func removeTorrent() string {
	return "not implemented"
}
