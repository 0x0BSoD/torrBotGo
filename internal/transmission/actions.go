package transmission

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/torrBotGo/internal/ctx"
	"github.com/0x0BSoD/transmission"
	"github.com/jackpal/bencode-go"
)

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
