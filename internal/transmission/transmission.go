package transmission

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/jackpal/bencode-go"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/transmission"

	"github.com/0x0BSoD/torrBotGo/internal/ctx"
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
	globalContext ctx.GlobalContext
}

func (c *Client) SendStatus() (string, error) {
	stats, err := c.globalContext.TrAPI.Session.Stats()
	if err != nil {
		return "", err
	}

	t, err := template.ParseFiles(c.globalContext.Cwd + "templates/status.gotmpl")
	if err != nil {
		return "", err
	}

	freeSpaceData, err := c.globalContext.TrAPI.FreeSpace(c.globalContext.TrAPI.Session.DownloadDir)
	if err != nil {
		return "", fmt.Errorf("error with %s: %s", c.globalContext.TrAPI.Session.DownloadDir, err)
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
	err := c.globalContext.TrAPI.Session.Update()
	if err != nil {
		return "", err
	}

	sc := c.globalContext.TrAPI.Session

	t, err := template.ParseFiles(c.globalContext.Cwd + "templates/config.gotmpl")
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
	err := c.globalContext.TrAPI.Session.Update()
	if err != nil {
		return err
	}

	sc := c.globalContext.TrAPI.Session

	b, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}

	err = sendNewMessage(c.globalContext.ChatID, fmt.Sprintf("`%s`", string(b)), nil)
	if err != nil {
		return err
	}

	return nil
}

//======================================================================================================================
// GET
//======================================================================================================================

func (c *Client) SendTorrent(id int64, torr *transmission.Torrent) error {
	t, err := template.ParseFiles(c.globalContext.Cwd + "templates/torrentListItem.gotmpl")
	if err != nil {
		return err
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
		return err
	}

	replyMarkup := torrentKbd(torr.HashString)
	err = sendNewMessage(id, dRes.String(), &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SendTorrentList(sf showFilter) error {
	tMap, err := c.globalContext.TorrentCache.Snapshot()
	if len(err) > 1 {
		// TODO: refactor it
		return fmt.Errorf("%v", err)
	}

	keys := make([]string, 0, len(tMap))
	for key := range tMap {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		one := tMap[keys[i]]
		two := tMap[keys[j]]
		return one.QueuePosition < two.QueuePosition
	})

	for _, i := range tMap {
		switch sf {
		case all:
			err := c.SendTorrent(c.globalContext.ChatID, i)
			if err != nil {
				return err
			}
		case active:
			if i.Status != 0 && i.ErrorString == "" {
				err := c.SendTorrent(c.globalContext.ChatID, i)
				if err != nil {
					return err
				}
			}
		case notActive:
			if i.Status == 0 {
				err := c.SendTorrent(c.globalContext.ChatID, i)
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

		t, err := template.ParseFiles(ctx.wd + "templates/torrent.gotmpl")
		if err != nil {
			log.Panic(err)
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

type filesList struct {
	Name        string
	Size        string
	Downloading bool
}

func sendTorrentDetails(hash string, messageID int, md5SumOld string) error {
	updateCache()

	t, err := getTorrentDetails(hash)
	if err != nil {
		return err
	}

	tHash := strings.ReplaceAll(strings.ReplaceAll(t, "`", ""), "\n", "")

	if glh.GetMD5Hash(tHash) == md5SumOld {
		return nil
	}

	replyMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendEditedMessage(ctx.chatID, messageID, t, &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func sendTorrentDetailsByID(torrentID int64) error {
	hash := ctx.TorrentCache.getHash(int(torrentID))

	t, err := getTorrentDetails(hash)
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
	files := *ctx.TorrentCache.Items[hash].Files
	filesStats := *ctx.TorrentCache.Items[hash].FileStats

	for i := 0; i < len(files); i++ {

		msg := tgbotapi.NewMessage(ctx.chatID, "")
		msg.ParseMode = "MarkdownV2"

		t, err := template.ParseFiles(ctx.wd + "templates/torrentFileItem.gotmpl")
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

func searchTorrent(text string) error {
	searchString := strings.Split(text, "t:")

	fmt.Println(searchString)

	if len(searchString) <= 1 {
		return errors.New("search string empty")
	}

	re := regexp.MustCompile(searchString[1])

	for _, t := range ctx.TorrentCache.Items {
		if re.Match([]byte(t.Name)) {
			err := sendTorrent(ctx.chatID, t)
			if err != nil {
				return err
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

func addTorrentMagnet(operation string) error {
	if operation == "add-no" {
		err := sendNewMessage(ctx.chatID, "Okay", nil)
		if err != nil {
			return err
		}
		return nil
	}

	path := ctx.TrAPI.Session.DownloadDir + strings.Split(operation, "-")[1]

	res, err := ctx.TrAPI.AddTorrent(transmission.AddTorrentArg{
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

	updateCache()

	return nil
}

func stopTorrent(hash string, messageID int, md5SumOld string) error {
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

	time.Sleep(6 * time.Second)

	updateCache()

	t, err := getTorrentDetails(hash)
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

func startTorrent(hash string, messageID int, md5SumOld string) error {
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

	time.Sleep(6 * time.Second)

	updateCache()

	t, err := getTorrentDetails(hash)
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
		err := ctx.TrAPI.RemoveTorrents([]*transmission.Torrent{TORRENT}, false)
		if err != nil {
			return err
		}
	case "yes+data":
		err := ctx.TrAPI.RemoveTorrents([]*transmission.Torrent{TORRENT}, true)
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
		err := ctx.TrAPI.QueueMoveTop([]*transmission.Torrent{TORRENT})
		if err != nil {
			return err
		}
	case "up":
		err := ctx.TrAPI.QueueMoveUp([]*transmission.Torrent{TORRENT})
		if err != nil {
			return err
		}
	case "down":
		err := ctx.TrAPI.QueueMoveDown([]*transmission.Torrent{TORRENT})
		if err != nil {
			return err
		}
	case "bottom":
		err := ctx.TrAPI.QueueMoveBottom([]*transmission.Torrent{TORRENT})
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

func addTorrentFile(operation string) error {
	if operation == "file+add-no" {
		TFILE = nil
		err := sendNewMessage(ctx.chatID, "Okay", nil)
		if err != nil {
			return err
		}
		return nil
	}

	path := ctx.TrAPI.Session.DownloadDir + strings.Split(operation, "-")[1]

	base64Str := base64.StdEncoding.EncodeToString(TFILE)

	res, err := ctx.TrAPI.AddTorrent(transmission.AddTorrentArg{
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
