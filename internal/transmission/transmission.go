package transmission

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"go.uber.org/zap"

	glh "github.com/0x0BSoD/goLittleHelpers"
	"github.com/0x0BSoD/transmission"

	"github.com/0x0BSoD/torrBotGo/internal/cache"
)

const (
	all showFilter = iota
	active
	notActive
)

// TORRENT - selected torrent
var TORRENT *transmission.Torrent

// Magnet - magnet link
var MAGNET string

// TFILE - downloaded torrent file
var TFILE []byte

// MESSAGEID - id of 'dialog' message
var MESSAGEID int

// Status - struct for storing current status of Transmission
type Status struct {
	Active     int
	Paused     int
	UploadS    string
	DownloadS  string
	Downloaded string
	Uploaded   string
	FreeSpace  string
}

type SessConfig struct {
	DownloadDir   string
	StartAdded    bool
	SpeedLimitD   string
	SpeedLimitDEn bool
	SpeedLimitU   string
	SpeedLimitUEn bool
	DownloadQEn   bool
	DownloadQSize int
}

type torrent struct {
	ID             int
	Peers          int
	Downloading    bool
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
	Percents       string
}

type showFilter int

type Client struct {
	transmission *transmission.Client
	log          *zap.Logger
	cwd          string
	cache        cache.Torrents
	chatID       int64
}

func (c *Client) SendStatus() (string, error) {
	stats, err := c.transmission.Session.Stats()
	if err != nil {
		return "", err
	}

	t, err := template.ParseFiles(c.cwd + "templates/status.gotmpl")
	if err != nil {
		return "", err
	}

	freeSpaceData, err := c.transmission.FreeSpace(c.transmission.Session.DownloadDir)
	if err != nil {
		return "", fmt.Errorf("error with %s: %s", c.transmission.Session.DownloadDir, err)
	}

	var dRes bytes.Buffer
	err = t.Execute(&dRes, Status{
		Active:     stats.ActiveTorrentCount,
		Paused:     stats.PausedTorrentCount,
		UploadS:    glh.ConvertBytes(float64(stats.UploadSpeed), glh.Speed),
		DownloadS:  glh.ConvertBytes(float64(stats.DownloadSpeed), glh.Speed),
		Uploaded:   glh.ConvertBytes(float64(stats.CurrentStats.UploadedBytes), glh.Size),
		Downloaded: glh.ConvertBytes(float64(stats.CurrentStats.DownloadedBytes), glh.Size),
		FreeSpace:  glh.ConvertBytes(float64(freeSpaceData), glh.Size),
	})
	if err != nil {
		return "", err
	}

	return dRes.String(), nil
}

func (c *Client) SendConfig() (string, error) {
	err := c.transmission.Session.Update()
	if err != nil {
		return "", err
	}

	sc := c.transmission.Session

	t, err := template.ParseFiles(c.cwd + "templates/config.gotmpl")
	if err != nil {
		return "", err
	}

	var dRes bytes.Buffer
	err = t.Execute(&dRes, SessConfig{
		DownloadDir:   sc.DownloadDir,
		StartAdded:    sc.StartAddedTorrents,
		SpeedLimitD:   glh.ConvertBytes(float64(sc.SpeedLimitDown), glh.Speed),
		SpeedLimitDEn: sc.SpeedLimitDownEnabled,
		SpeedLimitU:   glh.ConvertBytes(float64(sc.SpeedLimitUp), glh.Speed),
		SpeedLimitUEn: sc.SpeedLimitUpEnabled,
		DownloadQEn:   sc.DownloadQueueEnabled,
		DownloadQSize: sc.DownloadQueueSize,
	})
	if err != nil {
		return "", err
	}

	return dRes.String(), nil
}

func (c *Client) SendJSONConfig() error {
	err := c.transmission.Session.Update()
	if err != nil {
		return err
	}

	sc := c.transmission.Session

	b, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}

	err = sendNewMessage(c.chatID, fmt.Sprintf("`%s`", string(b)), nil)
	if err != nil {
		return err
	}

	return nil
}
