package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"
	"github.com/jackpal/bencode-go"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"text/template"
)

var TORRENT *transmission.Torrent
var MAGENT string
var TFILE []byte

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

func sendJsonConfig() (string, error) {
	err := ctx.TrApi.Session.Update()
	if err != nil {
		return "", err
	}

	sc := ctx.TrApi.Session

	if ctx.Debug {
		_ = glh.PrettyPrint(sc)
	}

	b, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		log.Panic(err)
	}

	return "```" + string(b) + "```", nil
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

func sendTorrentList(chatID int64, sf showFilter) error {
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
			err := sendTorrent(chatID, i)
			if err != nil {
				return err
			}
		case Active:
			if i.Status != 0 && i.ErrorString == "" {
				err := sendTorrent(chatID, i)
				if err != nil {
					return err
				}
			}
		case NotActive:
			if i.Status == 0 {
				err := sendTorrent(chatID, i)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getTorrentDetails(hash string) (string, error) {
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

type filesList struct {
	Name        string
	Size        string
	Downloading bool
}

func sendTorrentDetails(hash string, chatID int64, messageID int, md5SumOld string) error {
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
	err = sendEditedMessage(chatID, messageID, t, &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func sendTorrentFiles(hash string, chatID int64) error {
	tMap, err := ctx.TrApi.GetTorrentMap()
	if err != nil {
		return err
	}

	files := *tMap[hash].Files
	filesStats := *tMap[hash].FileStats

	for i := 0; i < len(files); i++ {

		msg := tgbotapi.NewMessage(chatID, "")
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

func addTorrentMagnetQuestion(chatID int64, text string) error {
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

	message := fmt.Sprintf("❔ To add:```%s```\nTrackers:```%s```", name, strings.Join(trackers, "\n"))

	kbd := torrentAddKbd(false)
	err := sendNewMessage(chatID, message, &kbd)
	if err != nil {
		return err
	}

	MAGENT = text

	return nil
}

func addTorrentMagnet(chatID int64, operation string) error {
	if operation == "add-no" {
		err := sendNewMessage(chatID, "Okay", nil)
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

	msg := fmt.Sprintf("`%s` - Sucesefully added", res.Name)

	err = sendNewMessage(chatID, msg, nil)
	if err != nil {
		return err
	}

	return nil
}

func stopTorrent(hash string, chatID int64, messageID int) error {
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
	err = sendEditedMessage(chatID, messageID, msgTxt, &newMarkup)
	if err != nil {
		return err
	}

	return nil
}

func startTorrent(hash string, chatID int64, messageID int) error {
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
	err = sendEditedMessage(chatID, messageID, msgTxt, &newMarkup)
	if err != nil {
		return err
	}

	return nil
}

func removeTorrentQuestion(hash string, chatID int64, messageID int) error {
	msgTxt, err := getTorrentDetails(hash)
	if err != nil {
		return err
	}

	replyMarkup := torrentDeleteKbd(hash)
	err = sendEditedMessage(chatID, messageID, msgTxt, &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func removeTorrent(hash string, chatID int64, messageID int, what string) error {
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
		err = sendEditedMessage(chatID, messageID, msgTxt, &replyMarkup)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("nope, failed")
	}

	err := sendEditedMessage(chatID, messageID, "removed", nil)
	if err != nil {
		return err
	}

	return nil
}

func queueTorrentQuestion(hash string, chatID int64, messageID int) error {
	msgTxt, err := getTorrentDetails(hash)
	if err != nil {
		return err
	}

	replyMarkup := torrentQueueKbd(hash)
	err = sendEditedMessage(chatID, messageID, msgTxt, &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func queueTorrent(hash string, chatID int64, messageID int, what string) error {
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
	err = sendEditedMessage(chatID, messageID, msgTxt, &replyMarkup)
	if err != nil {
		return err
	}
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

func addTorrentFileQuestion(chatID int64, fileID string) error {
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

	freeSpaceData, err := ctx.TrApi.FreeSpace(ctx.TrApi.Session.DownloadDir)
	kbdAdd := torrentAddKbd(true)

	// torrent from rutracker
	if bto.Comment != "" {
		imgUrl, err := getImgFromTracker(bto.Comment)
		if err != nil {
			return err
		}

		_, err = url.ParseRequestURI(imgUrl)
		if err != nil {
			return err
		}

		client := httpClient()

		resp, err := client.Get(imgUrl)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		err = sendNewImagedMessage(chatID, bto.Info.Name, resp.Body, &kbdAdd)
		if err != nil {
			return err
		}

	} else {
		err = sendNewMessage(chatID, "```"+"Free space: "+glh.ConvertBytes(float64(freeSpaceData), glh.Size)+"```", nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func addTorrentFile(chatID int64, operation string) error {

	if operation == "file+add-no" {
		TFILE = nil
		err := sendNewMessage(chatID, "Okay", nil)
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

	msg := fmt.Sprintf("`%s` - Sucesefully added", res.Name)

	err = sendNewMessage(chatID, msg, nil)
	if err != nil {
		return err
	}

	return nil
}
