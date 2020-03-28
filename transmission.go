package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
	"github.com/jackpal/bencode-go"
)

// TORRENT - selected torrent
var TORRENT *transmission.Torrent

// MAGENT - magent link
var MAGENT string

// TFILE - downloaded torrent file
var TFILE []byte

// MESSAGEID - id of 'dialog' message
var MESSAGEID int

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
	err = t.Execute(&dRes, status{
		Active:     stats.ActiveTorrentCount,
		Paused:     stats.PausedTorrentCount,
		UploadS:    glh.ConvertBytes(float64(stats.UploadSpeed), glh.Speed),
		DownloadS:  glh.ConvertBytes(float64(stats.DownloadSpeed), glh.Speed),
		Uploaded:   glh.ConvertBytes(float64(stats.CurrentStats.UploadedBytes), glh.Size),
		Downloaded: glh.ConvertBytes(float64(stats.CurrentStats.DownloadedBytes), glh.Size),
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
}

func sendConfig() (string, error) {
	err := ctx.TrApi.Session.Update()
	if err != nil {
		return "", err
	}

	sc := ctx.TrApi.Session

	t, err := template.ParseFiles("templates/config.gotmpl")
	if err != nil {
		return "", err
	}

	if ctx.Debug {
		_ = glh.PrettyPrint(sc)
	}

	var dRes bytes.Buffer
	err = t.Execute(&dRes, sessConfig{
		DownloadDir:   sc.DownloadDir,
		StartAdded:    sc.StartAddedTorrents,
		SpeedLimitD:   glh.ConvertBytes(float64(sc.SpeedLimitDown), glh.Speed),
		SpeedLimitDEn: sc.SpeedLimitDownEnabled,
		SpeedLimitU:   glh.ConvertBytes(float64(sc.SpeedLimitUp), glh.Speed),
		SpeedLimitUEn: sc.SpeedLimitUpEnabled,
	})
	if err != nil {
		return "", err
	}

	return dRes.String(), nil
}

func sendJSONConfig() error {
	err := ctx.TrApi.Session.Update()
	if err != nil {
		return err
	}

	sc := ctx.TrApi.Session

	if ctx.Debug {
		_ = glh.PrettyPrint(sc)
	}

	b, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}

	err = sendNewMessage(ctx.chatID, fmt.Sprintf("`%s`", string(b)), nil)
	if err != nil {
		return err
	}

	return nil
}

type Torrent struct {
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
		return err
	}

	replyMarkup := torrentKbd(torr.HashString)
	err = sendNewMessage(id, dRes.String(), &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func sendTorrentList(sf showFilter) error {
	// TODO: sort by priority
	//sort.Slice(torrents[:], func(i, j int) bool {
	//	return torrents[i].QueuePosition < torrents[i].QueuePosition
	//})

	ctx.Mutex.Lock()
	defer ctx.Mutex.Unlock()
	for _, i := range ctx.TorrentCache.Items {
		switch sf {
		case All:
			err := sendTorrent(ctx.chatID, i)
			if err != nil {
				return err
			}
		case Active:
			if i.Status != 0 && i.ErrorString == "" {
				err := sendTorrent(ctx.chatID, i)
				if err != nil {
					return err
				}
			}
		case NotActive:
			if i.Status == 0 {
				err := sendTorrent(ctx.chatID, i)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getTorrentDetails(hash string) (string, error) {
	var ok bool
	if TORRENT, ok = ctx.TorrentCache.Items[hash]; ok {
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
		err = t.Execute(&dRes, Torrent{
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

type filesList struct {
	Name        string
	Size        string
	Downloading bool
}

func sendTorrentDetails(hash string, messageID int, md5SumOld string) error {
	updateCache()
	t, err := getTorrentDetails(hash)

	fmt.Println("====")
	fmt.Println(strings.ReplaceAll(t, "`", ""))
	fmt.Println("====")

	if glh.GetMD5Hash(strings.ReplaceAll(t, "`", "")) == md5SumOld {
		return nil
	}

	if err != nil {
		return err
	}
	replyMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendEditedMessage(ctx.chatID, messageID, t, &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func sendTorrentDetailsByID(torrentID int64) error {
	hash := ctx.TorrentCache.GetHash(int(torrentID))

	t, err := getTorrentDetails(hash)

	fmt.Println("====")
	fmt.Println(strings.ReplaceAll(t, "`", ""))
	fmt.Println("====")

	if err != nil {
		return err
	}
	replyMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendNewMessage(ctx.chatID, t, &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func sendTorrentFiles(hash string) error {
	//tMap, err := ctx.TrApi.GetTorrentMap()
	//if err != nil {
	//	return err
	//}

	files := *ctx.TorrentCache.Items[hash].Files
	filesStats := *ctx.TorrentCache.Items[hash].FileStats

	for i := 0; i < len(files); i++ {

		msg := tgbotapi.NewMessage(ctx.chatID, "")
		msg.ParseMode = "MarkdownV2"

		t, err := template.ParseFiles("templates/torrentFile.gotmpl")
		if err != nil {
			log.Panic(err)
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
				log.Panic(err)
			}
		}
	}

	return nil
}

//======================================================================================================================
// ACTIONS WITH TORRENT
//======================================================================================================================

func addTorrentMagnetQuestion(text string, messageID int) error {
	var name string
	var trackers []string
	for _, i := range strings.Split(text, "&") {
		decoded, err := url.QueryUnescape(i)
		if err != nil {
			panic(err)
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

	MAGENT = text
	MESSAGEID = messageID

	return nil
}

func addTorrentMagnet(operation string) error {
	if operation == "add-no" {
		err := sendNewMessage(ctx.chatID, "Okay", nil)
		if err != nil {
			return err
		}
		return nil
	}

	path := ctx.TrApi.Session.DownloadDir + strings.Split(operation, "-")[1]

	res, err := ctx.TrApi.AddTorrent(transmission.AddTorrentArg{
		DownloadDir: path,
		Filename:    MAGENT,
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

	updateCache()

	return nil
}

func stopTorrent(hash string, messageID int) error {
	if TORRENT == nil {
		_, err := getTorrentDetails(hash)
		if err != nil {
			return err
		}
	}

	err := TORRENT.Stop()
	if err != nil {
		return err
	}
	TORRENT = nil

	msgTxt, err := getTorrentDetails(hash)
	newMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendEditedMessage(ctx.chatID, messageID, msgTxt, &newMarkup)
	if err != nil {
		return err
	}

	updateCache()

	return nil
}

func startTorrent(hash string, messageID int) error {
	if TORRENT == nil {
		_, err := getTorrentDetails(hash)
		if err != nil {
			return err
		}
	}

	err := TORRENT.Start()
	if err != nil {
		return err
	}
	TORRENT = nil

	msgTxt, err := getTorrentDetails(hash)
	if err != nil {
		return err
	}

	newMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendEditedMessage(ctx.chatID, messageID, msgTxt, &newMarkup)
	if err != nil {
		return err
	}

	updateCache()

	return nil
}

func removeTorrentQuestion(hash string, messageID int) error {
	msgTxt, err := getTorrentDetails(hash)
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

func removeTorrent(hash string, messageID int, what string) error {
	whatS := strings.Split(what, "-")[1]
	switch whatS {
	case "yes":
		err := ctx.TrApi.RemoveTorrents([]*transmission.Torrent{TORRENT}, false)
		if err != nil {
			return err
		}
	case "yes+data":
		err := ctx.TrApi.RemoveTorrents([]*transmission.Torrent{TORRENT}, true)
		if err != nil {
			return err
		}
	case "no":
		msgTxt, err := getTorrentDetails(hash)
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

	updateCache()

	return nil
}

func queueTorrentQuestion(hash string, messageID int) error {
	msgTxt, err := getTorrentDetails(hash)
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

func queueTorrent(hash string, messageID int, what string) error {
	whatS := strings.Split(what, "-")[1]
	switch whatS {
	case "top":
		err := ctx.TrApi.QueueMoveTop([]*transmission.Torrent{TORRENT})
		if err != nil {
			return err
		}
	case "up":
		err := ctx.TrApi.QueueMoveUp([]*transmission.Torrent{TORRENT})
		if err != nil {
			return err
		}
	case "down":
		err := ctx.TrApi.QueueMoveDown([]*transmission.Torrent{TORRENT})
		if err != nil {
			return err
		}
	case "bottom":
		err := ctx.TrApi.QueueMoveBottom([]*transmission.Torrent{TORRENT})
		if err != nil {
			return err
		}
	case "no":
		// pass
	default:
		return fmt.Errorf("nope, failed")
	}

	msgTxt, err := getTorrentDetails(hash)
	if err != nil {
		return err
	}

	replyMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendEditedMessage(ctx.chatID, messageID, msgTxt, &replyMarkup)
	if err != nil {
		return err
	}

	updateCache()

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

func addTorrentFileQuestion(fileID string, messageID int) error {
	_url, err := ctx.Bot.GetFileDirectURL(fileID)
	if err != nil {
		return err
	}

	resp, err := http.Get(_url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	TFILE, err = ioutil.ReadAll(resp.Body)
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

	// freeSpaceData, err := ctx.TrApi.FreeSpace(ctx.TrApi.Session.DownloadDir)
	kbdAdd := torrentAddKbd(true)

	fallback := func(name string, kbd *tgbotapi.InlineKeyboardMarkup) error {
		err := sendNewMessage(ctx.chatID, name, kbd)
		if err != nil {
			return err
		}
		return nil
	}

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

	MESSAGEID = messageID

	return nil
}

func addTorrentFile(operation string) error {

	if operation == "file+add-no" {
		TFILE = nil
		err := sendNewMessage(ctx.chatID, "Okay", nil)
		if err != nil {
			return err
		}
		return nil
	}

	path := ctx.TrApi.Session.DownloadDir + strings.Split(operation, "-")[1]

	base64Str := base64.StdEncoding.EncodeToString(TFILE)

	res, err := ctx.TrApi.AddTorrent(transmission.AddTorrentArg{
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

	updateCache()

	return nil
}

func updateCache() {
	tMap, err := ctx.TrApi.GetTorrentMap()
	if err != nil {
		panic(err)
	}
	changed := ctx.TorrentCache.Update(tMap)

	if len(changed) == 0 {
		return
	}

	for _, i := range changed {
		if i.ErrorString != "" {
			err := sendNewMessage(ctx.chatID, fmt.Sprintf("🔥️ Failed\n%s\nError:\n%s", i.Name, i.ErrorString), nil)
			if err != nil {
				panic(err)
			}
		} else if i.Status != 4 {
			err := sendNewMessage(ctx.chatID, fmt.Sprintf("🎉 Downloaded\n%s", i.Name), nil)
			if err != nil {
				panic(err)
			}
		}
	}
}
