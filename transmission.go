package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"text/template"
	"time"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/torrBotGo/internal/cache"
	"github.com/0x0BSoD/transmission"
	"github.com/jackpal/bencode-go"
	"go.uber.org/zap"
)

// TORRENT - selected torrent
var TORRENT *transmission.Torrent

// MAGNET - magnet link
var MAGNET string

// TFILE - downloaded torrent file
var TFILE []byte

// MESSAGEID - id of 'dialog' message
var MESSAGEID int

type trClient struct {
	API *transmission.Client
	log *zap.Logger
	cwd string
}

func trInit(cfg *config, log *zap.Logger) (*trClient, error) {
	conf := transmission.Config{
		Address:  cfg.Transmission.Config.URI,
		User:     cfg.Transmission.Config.User,
		Password: cfg.Transmission.Config.Password,
	}

	t, err := transmission.New(conf)
	t.Context = context.TODO()
	if err != nil {
		return nil, err
	}

	if (transmission.SetSessionArgs{}) != cfg.Transmission.Custom {
		log.Info("setting custom transmission parameters")
		err := t.Session.Set(cfg.Transmission.Custom)
		if err != nil {
			log.Sugar().Errorf("getting tg updates failed: %w", err)
			return nil, err
		}
	}

	log.Info("updating transmission session info ")
	err = t.Session.Update()
	if err != nil {
		return nil, err
	}

	tMap, err := t.GetTorrentMap()
	if err != nil {
		return nil, err
	}
	log.Info("setting torrents cache")
	ctx.TorrentCache = cache.New(tMap)

	var result trClient
	result.API = t
	result.log = log
	result.cwd = cfg.App.WorkingDir

	return &result, nil
}

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
	FreeSpace  string
}

func (c *trClient) GetStatus() (string, error) {
	stats, err := c.API.Session.Stats()
	if err != nil {
		return "", err
	}

	t, err := template.ParseFiles(ctx.wd + "templates/status.gotmpl")
	if err != nil {
		return "", err
	}

	freeSpaceData, err := c.API.FreeSpace(c.API.Session.DownloadDir)
	if err != nil {
		return "", err
	}

	var dRes bytes.Buffer
	err = t.Execute(&dRes, status{
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

type sessConfig struct {
	DownloadDir   string
	StartAdded    bool
	SpeedLimitD   string
	SpeedLimitDEn bool
	SpeedLimitU   string
	SpeedLimitUEn bool
	DownloadQEn   bool
	DownloadQSize int
}

func (c *trClient) GetConfig() (string, error) {
	err := c.API.Session.Update()
	if err != nil {
		return "", err
	}

	sc := c.API.Session

	t, err := template.ParseFiles(ctx.wd + "templates/config.gotmpl")
	if err != nil {
		return "", err
	}

	var dRes bytes.Buffer
	err = t.Execute(&dRes, sessConfig{
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

func (c *trClient) GetJSONConfig() (string, error) {
	err := c.API.Session.Update()
	if err != nil {
		return "", err
	}

	sc := c.API.Session

	if ctx.Debug {
		_ = glh.PrettyPrint(sc)
	}

	b, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("`%s`", string(b)), nil
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

const (
	all showFilter = iota
	active
	notActive
)

type filesList struct {
	Name        string
	Size        string
	Downloading bool
}

//======================================================================================================================
// HELPERS
//======================================================================================================================

func (c *trClient) renderTorrent(torr *transmission.Torrent) (string, string, error) {
	t, err := template.ParseFiles(ctx.wd + "templates/torrentListItem.gotmpl")
	if err != nil {
		return "", "", err
	}

	icon, status := parseStatus(torr.Status)
	var dRes bytes.Buffer
	err = t.Execute(&dRes, torrent{
		ID:          torr.ID,
		Name:        torr.Name,
		Status:      status,
		Icon:        icon,
		ErrorString: torr.ErrorString,
		Comment:     torr.Comment,
		Hash:        torr.HashString,
		PosInQ:      torr.QueuePosition,
	})
	if err != nil {
		return "", "", err
	}

	return torr.HashString, dRes.String(), nil
}

func (c *trClient) details(hash string) (string, error) {
	var ok bool
	if TORRENT, ok = ctx.TorrentCache.GetByHash(hash); ok {
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

		t, err := template.ParseFiles(ctx.wd + "templates/torrent.gotmpl")
		if err != nil {
			return "", err
		}

		var dRes bytes.Buffer
		err = t.Execute(&dRes, torrent{
			ID:             TORRENT.ID,
			Peers:          len(*TORRENT.Peers),
			Downloading:    TORRENT.Status == 4,
			PosInQ:         TORRENT.QueuePosition,
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
			Percents:       fmt.Sprintf("%.2f%%", TORRENT.PercentDone*100.0),
		})
		if err != nil {
			return "", err
		}

		return dRes.String(), nil
	}

	return "", errors.New("torrent not found")
}

//======================================================================================================================
// GET
//======================================================================================================================

func (c *trClient) Torrents(sf showFilter) (map[string]string, error) {
	items, _ := ctx.TorrentCache.Snapshot()

	result := make(map[string]string)

	for _, i := range items {
		switch sf {
		case all:
			hash, text, err := c.renderTorrent(i)
			if err != nil {
				return nil, err
			}
			result[hash] = text
		case active:
			if i.Status != transmission.StatusStopped && i.ErrorString == "" {
				hash, text, err := c.renderTorrent(i)
				if err != nil {
					return nil, err
				}
				result[hash] = text
			}
		case notActive:
			if i.Status == transmission.StatusStopped {
				hash, text, err := c.renderTorrent(i)
				if err != nil {
					return nil, err
				}
				result[hash] = text
			}
		}
	}

	return result, nil
}

func (c *trClient) TorrentDetails(hash string, messageID int, md5SumOld string) (string, error) {
	c.updateCache(context.TODO(), &ctx)

	t, err := c.details(hash)
	if err != nil {
		return "", err
	}

	tHash := glh.GetMD5Hash(strings.ReplaceAll(strings.ReplaceAll(t, "`", ""), "\n", ""))
	if tHash == md5SumOld {
		return "", nil
	}

	return t, nil
}

func (c *trClient) TorrentDetailsByID(torrentID int64) (string, string, error) {
	hash, _ := ctx.TorrentCache.GetHash((int(torrentID)))
	t, err := c.details(hash)
	if err != nil {
		return "", "", err
	}

	return hash, t, nil
}

func (c *trClient) sendTorrentFiles(hash string) error {
	t, _ := ctx.TorrentCache.GetByHash(hash)
	files := *t.Files
	filesStats := *t.FileStats

	for i := range len(files) {
		msg := tgbotapi.NewMessage(ctx.chatID, "")
		msg.ParseMode = "MarkdownV2"

		t, err := template.ParseFiles(ctx.wd + "templates/torrentFileItem.gotmpl")
		if err != nil {
			return err
		}

		var dRes bytes.Buffer
		err = t.Execute(&dRes, filesList{
			Name:        files[i].Name,
			Size:        glh.ConvertBytes(float64(files[i].Length), glh.Size),
			Downloading: filesStats[i].Wanted,
		})
		if err != nil {
			return err
		}
		msg.Text = dRes.String()

		if msg.Text != "" {
			if _, err := ctx.Bot.Send(msg); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *trClient) SearchTorrent(text string) (map[string]string, error) {
	searchString := strings.Split(text, "t:")

	fmt.Println(searchString)

	if len(searchString) <= 1 {
		return nil, errors.New("search string empty")
	}

	re := regexp.MustCompile(searchString[1])

	items, _ := ctx.TorrentCache.Snapshot()
	result := make(map[string]string)
	for _, t := range items {
		if re.Match([]byte(t.Name)) {
			hash, text, err := c.renderTorrent(t)
			if err != nil {
				return nil, err
			}
			result[hash] = text
		}
	}

	return result, nil
}

//======================================================================================================================
// ACTIONS WITH TORRENT
//======================================================================================================================

func (c *trClient) addTorrentMagnetQuestion(text string, messageID int) error {
	var name string
	var trackers []string
	for i := range strings.SplitSeq(text, "&") {
		decoded, err := url.QueryUnescape(i)
		if err != nil {
			return err
		}

		if strings.HasPrefix(decoded, "dn=") {
			name = strings.ReplaceAll(decoded, "dn=", "")
		}
		if strings.HasPrefix(decoded, "tr=") {
			trackers = append(trackers, strings.ReplaceAll(decoded, "tr=", ""))
		}
	}

	message := fmt.Sprintf("`%s`\nTrackers:`%s`", name, strings.Join(trackers, "\n"))

	kbd := torrentAddKbd(false)
	err := sendNewMessage(ctx.chatID, message, &kbd)
	if err != nil {
		return err
	}

	MAGNET = text
	MESSAGEID = messageID

	return nil
}

func (c *trClient) addTorrentMagnet(operation string) error {
	if operation == "add-no" {
		err := sendNewMessage(ctx.chatID, "Okay", nil)
		if err != nil {
			return err
		}
		return nil
	}

	path := c.API.Session.DownloadDir + strings.Split(operation, "-")[1]

	res, err := c.API.AddTorrent(transmission.AddTorrentArg{
		DownloadDir: path,
		Filename:    MAGNET,
		Paused:      false,
	})
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Sucesefully added\n`%s`\nID:`%d`", res.Name, res.ID)

	err = removeMessage(ctx.chatID, MESSAGEID)
	if err != nil {
		return err
	}

	err = sendNewMessage(ctx.chatID, msg, nil)
	if err != nil {
		return err
	}

	c.updateCache(context.TODO(), &ctx)

	return nil
}

func (c *trClient) stopTorrent(hash string, messageID int, md5SumOld string) error {
	if TORRENT == nil {
		_, err := c.details(hash)
		if err != nil {
			return err
		}
	}

	err := TORRENT.Stop()
	if err != nil {
		return err
	}
	TORRENT = nil

	time.Sleep(6 * time.Second)

	c.updateCache(context.TODO(), &ctx)

	t, err := c.details(hash)
	if err != nil {
		return err
	}

	tHash := strings.ReplaceAll(strings.ReplaceAll(t, "`", ""), "\n", "")

	if glh.GetMD5Hash(tHash) == md5SumOld {
		return nil
	}

	newMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendEditedMessage(ctx.chatID, messageID, t, &newMarkup)
	if err != nil {
		return err
	}

	return nil
}

func (c *trClient) startTorrent(hash string, messageID int, md5SumOld string) error {
	if TORRENT == nil {
		_, err := c.details(hash)
		if err != nil {
			return err
		}
	}

	err := TORRENT.Start()
	if err != nil {
		return err
	}
	TORRENT = nil

	time.Sleep(6 * time.Second)

	c.updateCache(context.TODO(), &ctx)

	t, err := c.details(hash)
	if err != nil {
		return err
	}

	tHash := strings.ReplaceAll(strings.ReplaceAll(t, "`", ""), "\n", "")

	if glh.GetMD5Hash(tHash) == md5SumOld {
		return nil
	}

	newMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendEditedMessage(ctx.chatID, messageID, t, &newMarkup)
	if err != nil {
		return err
	}

	return nil
}

func (c *trClient) removeTorrentQuestion(hash string, messageID int) error {
	msgTxt, err := c.details(hash)
	if err != nil {
		return err
	}

	replyMarkup := torrentDeleteKbd(hash)
	err = sendEditedMessage(ctx.chatID, messageID, msgTxt, &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func (c *trClient) removeTorrent(hash string, messageID int, what string) error {
	whatS := strings.Split(what, "-")[1]
	switch whatS {
	case "yes":
		err := c.API.RemoveTorrents([]*transmission.Torrent{TORRENT}, false)
		if err != nil {
			return err
		}
	case "yes+data":
		err := c.API.RemoveTorrents([]*transmission.Torrent{TORRENT}, true)
		if err != nil {
			return err
		}
	case "no":
		msgTxt, err := c.details(hash)
		if err != nil {
			return err
		}

		replyMarkup := torrentDetailKbd(hash, TORRENT.Status)
		err = sendEditedMessage(ctx.chatID, messageID, msgTxt, &replyMarkup)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("nope, failed")
	}

	err := sendEditedMessage(ctx.chatID, messageID, "removed", nil)
	if err != nil {
		return err
	}

	c.updateCache(context.TODO(), &ctx)

	return nil
}

func (c *trClient) queueTorrentQuestion(hash string, messageID int) error {
	msgTxt, err := c.details(hash)
	if err != nil {
		return err
	}

	replyMarkup := torrentQueueKbd(hash)
	err = sendEditedMessage(ctx.chatID, messageID, msgTxt, &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func (c *trClient) queueTorrent(hash string, messageID int, what string) error {
	whatS := strings.Split(what, "-")[1]
	switch whatS {
	case "top":
		err := c.API.QueueMoveTop([]*transmission.Torrent{TORRENT})
		if err != nil {
			return err
		}
	case "up":
		err := c.API.QueueMoveUp([]*transmission.Torrent{TORRENT})
		if err != nil {
			return err
		}
	case "down":
		err := c.API.QueueMoveDown([]*transmission.Torrent{TORRENT})
		if err != nil {
			return err
		}
	case "bottom":
		err := c.API.QueueMoveBottom([]*transmission.Torrent{TORRENT})
		if err != nil {
			return err
		}
	case "no":
		// pass
	default:
		return fmt.Errorf("nope, failed")
	}

	msgTxt, err := c.details(hash)
	if err != nil {
		return err
	}

	replyMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendEditedMessage(ctx.chatID, messageID, msgTxt, &replyMarkup)
	if err != nil {
		return err
	}

	c.updateCache(context.TODO(), &ctx)

	return nil
}

type bencodeInfo struct {
	Length int    `bencode:"length"`
	Name   string `bencode:"name"`
}

type bencodeTorrent struct {
	Comment  string      `bencode:"comment"`
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

func (c *trClient) addTorrentFileQuestion(fileID string, messageID int) error {
	_url, err := ctx.Bot.GetFileDirectURL(fileID)
	if err != nil {
		return err
	}

	resp, err := http.Get(_url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	TFILE, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(bytes.NewReader(TFILE), &bto)
	if err != nil {
		return err
	}

	if ctx.Debug {
		_ = glh.PrettyPrint(bto)
	}

	kbdAdd := torrentAddKbd(true)

	fallback := func(name string, kbd *tgbotapi.InlineKeyboardMarkup) error {
		err := sendNewMessage(ctx.chatID, name, kbd)
		if err != nil {
			return err
		}
		return nil
	}

	MESSAGEID = messageID

	// torrent from rutracker
	if bto.Comment != "" {
		imgURL, err := getImgFromTrackerRutracker(bto.Comment)
		if err != nil {
			err := fallback(bto.Info.Name, &kbdAdd)
			if err != nil {
				return err
			}
			return nil
		}

		_, err = url.ParseRequestURI(imgURL)
		if err != nil {
			err := fallback(bto.Info.Name, &kbdAdd)
			if err != nil {
				return err
			}
			return nil
		}

		client := httpClient()

		resp, err := client.Get(imgURL)
		if err != nil {
			err := fallback(bto.Info.Name, &kbdAdd)
			if err != nil {
				return err
			}
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			err := fallback(bto.Info.Name, &kbdAdd)
			if err != nil {
				return err
			}
			return nil
		}

		err = sendNewImagedMessage(ctx.chatID, bto.Info.Name, resp.Body, &kbdAdd)
		if err != nil {
			return err
		}

	} else {
		err := fallback(bto.Info.Name, &kbdAdd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *trClient) addTorrentFile(operation string) error {
	if operation == "file+add-no" {
		TFILE = nil
		err := sendNewMessage(ctx.chatID, "Okay", nil)
		if err != nil {
			return err
		}
		return nil
	}

	path := c.API.Session.DownloadDir + strings.Split(operation, "-")[1]

	base64Str := base64.StdEncoding.EncodeToString(TFILE)

	res, err := c.API.AddTorrent(transmission.AddTorrentArg{
		DownloadDir: path,
		Metainfo:    base64Str,
		Paused:      false,
	})
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Sucesefully added\n`%s`\nID:`%d`", res.Name, res.ID)

	err = removeMessage(ctx.chatID, MESSAGEID)
	if err != nil {
		return err
	}

	err = sendNewMessage(ctx.chatID, msg, nil)
	if err != nil {
		return err
	}

	c.updateCache(context.TODO(), &ctx)

	return nil
}

const doneEpsilon = 0.9999

// startCacheUpdater - watcher that copare current torrent state and stored in memory
func (c *trClient) startCacheUpdater(ctx context.Context, interval time.Duration, gCtx *GlobalContext) {
	t := time.NewTicker(interval)
	defer t.Stop()

	if err := c.updateCache(ctx, gCtx); err != nil {
		fmt.Printf("StartCacheUpdater: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := c.updateCache(ctx, gCtx); err != nil {
				fmt.Printf("StartCacheUpdater: %v", err)
			}
		}
	}
}

func (c *trClient) updateCache(ctx context.Context, gCtx *GlobalContext) error {
	tMap, err := c.API.GetTorrentMap()
	if err != nil {
		return fmt.Errorf("updateCache: fetch: %w", err)
	}

	changed := gCtx.TorrentCache.Update(tMap)
	if len(changed) == 0 {
		return nil
	}

	var msgs []string
	for _, t := range changed {
		if t == nil {
			continue
		}

		if t.ErrorString != "" {
			msgs = append(msgs,
				fmt.Sprintf("Failed\n%s\nError:\n%s", t.Name, t.ErrorString))
			continue
		}

		if t.PercentDone >= doneEpsilon && t.Status == transmission.StatusSeeding {
			msgs = append(msgs, fmt.Sprintf("Downloaded\n%s", t.Name))
		}
	}

	for _, m := range msgs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := sendNewMessage(gCtx.chatID, m, nil); err != nil {
			fmt.Printf("UpdateCache: send failed: %v", err)
		}
	}

	return nil
}
