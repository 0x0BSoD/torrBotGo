package transmission

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"text/template"

	glh "github.com/0x0BSoD/goLittleHelpers"
	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/0x0BSoD/torrBotGo/internal/ctx"
	"github.com/0x0BSoD/transmission"
)

func (c *Client) SendTorrent(id int64, torr *transmission.Torrent) error {
	t, err := template.ParseFiles(c.cwd + "templates/torrentListItem.gotmpl")
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
	tMap, err := c.cache.Snapshot()
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
			err := c.SendTorrent(c.chatID, i)
			if err != nil {
				return err
			}
		case active:
			if i.Status != 0 && i.ErrorString == "" {
				err := c.SendTorrent(c.chatID, i)
				if err != nil {
					return err
				}
			}
		case notActive:
			if i.Status == 0 {
				err := c.SendTorrent(c.chatID, i)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *Client) getTorrentDetails(hash string) (string, error) {
	var ok bool
	if TORRENT, ok = c.cache.GetByHash(hash); ok {
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

		t, err := template.ParseFiles(c.cwd + "templates/torrent.gotmpl")
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

func (c *Client) sendTorrentDetailsByID(torrentID int64) error {
	hash, _ := c.cache.GetHash(int(torrentID))

	t, err := getTorrentDetails(hash)
	if err != nil {
		return err
	}
	replyMarkup := torrentDetailKbd(hash, TORRENT.Status)
	err = sendNewMessage(c.chatID, t, &replyMarkup)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) sendTorrentFiles(hash string) error {
	t, _ := c.cache.GetByHash(hash)
	files := *t.Files
	filesStats := *t.FileStats

	for i := 0; i < len(files); i++ {

		msg := tgbotapi.NewMessage(c.chatID, "")
		msg.ParseMode = "MarkdownV2"

		t, err := template.ParseFiles(c.cwd + "templates/torrentFileItem.gotmpl")
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

func (c *Client) searchTorrent(text string) error {
	searchString := strings.Split(text, "t:")

	fmt.Println(searchString)

	if len(searchString) <= 1 {
		return errors.New("search string empty")
	}

	re := regexp.MustCompile(searchString[1])

	items, _ := c.cache.Snapshot()
	for _, t := range items {
		if re.Match([]byte(t.Name)) {
			err := sendTorrent(c.chatID, t)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
