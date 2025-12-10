package transmission

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackpal/bencode-go"

	glh "github.com/0x0BSoD/goLittleHelpers"
	"github.com/0x0BSoD/transmission"
)

type Torrent struct {
	ID             int
	Peers          int
	Downloading    bool
	Active         bool
	Name           string
	Status         string
	StatusCode     int
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

type TorrentFilesItem struct {
	Name        string
	Size        string
	Downloading bool
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

var (
	ErrorFilterNotFound  = errors.New("unknown filter")
	ErrorTorrentNotFound = errors.New("torrent not found")
)

func (c *Client) Torrents(showFilter string) (map[string]*transmission.Torrent, error) {
	items, _ := c.cache.Snapshot()

	result := make(map[string]*transmission.Torrent)

	for _, torrent := range items {
		switch showFilter {
		case "All torrents":
			result[torrent.HashString] = torrent
		case "Active torrents":
			if torrent.Status != transmission.StatusStopped && torrent.ErrorString == "" {
				result[torrent.HashString] = torrent
			}
		case "Not Active torrents":
			if torrent.Status == transmission.StatusStopped {
				result[torrent.HashString] = torrent
			}
		default:
			return result, ErrorFilterNotFound
		}
	}

	return result, nil
}

func (c *Client) AddByMagnetDialog(input string) (string, error) {
	var name string
	var trackers []string

	for i := range strings.SplitSeq(input, "&") {
		decoded, err := url.QueryUnescape(i)
		if err != nil {
			return "", err
		}

		if strings.HasPrefix(decoded, "dn=") {
			name = strings.ReplaceAll(decoded, "dn=", "")
		}
		if strings.HasPrefix(decoded, "tr=") {
			trackers = append(trackers, strings.ReplaceAll(decoded, "tr=", ""))
		}
	}

	c.storage.magentLink = input

	return fmt.Sprintf("`%s`\nTrackers:`%s`", name, strings.Join(trackers, "\n")), nil
}

func (c *Client) AddByMagent(operation string) (string, error) {
	if operation == "add-no" {
		c.storage.magentLink = ""
		return "Okay", nil
	}

	pathKey := strings.Split(operation, "-")[1]
	path := filepath.Join(c.API.Session.DownloadDir + c.Categories[pathKey].Path)

	res, err := c.API.AddTorrent(transmission.AddTorrentArg{
		DownloadDir: path,
		Filename:    c.storage.magentLink,
		Paused:      false,
	})
	if err != nil {
		return "", err
	}

	_ = c.updateCache(context.TODO())
	return fmt.Sprintf("Successfully added\n`%s`\nID:`%d`", res.Name, res.ID), nil
}

func (c *Client) AddByFileDialog(directURL string) (string, string, error) {
	resp, err := http.Get(directURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	c.storage.tFile, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(bytes.NewReader(c.storage.tFile), &bto)
	if err != nil {
		return "", "", err
	}

	// torrent from rutracker
	if bto.Comment != "" {
		if !strings.HasPrefix(bto.Comment, "https://rutracker.org/") {
			c.logger.Sugar().Debugf("not a RuTracker: %s", bto.Comment)
			return bto.Info.Name, "", nil
		}

		doc, err := fetchPage(bto.Comment)
		if err != nil {
			return bto.Info.Name, "", nil
		}

		imgURL := getImgURLRutracker(doc)
		category := getCategoryRutracker(doc)
		suggestedCat := matchCategory(category, c.Categories)

		resultText := fmt.Sprintf("%s::%s", suggestedCat, bto.Info.Name)

		_, err = url.ParseRequestURI(imgURL)
		if err != nil {
			return resultText, "", nil
		}

		client := httpClient()
		resp, err := client.Get(imgURL)
		if err != nil {
			return resultText, "", nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return resultText, "", nil
		}

		hasher := sha1.New()
		tmHash := strconv.Itoa(time.Now().Nanosecond())
		hasher.Write([]byte(tmHash))
		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		imgPath := filepath.Join(c.mediaPath, sha)
		file, err := os.Create(imgPath)
		if err != nil {
			return "", "", err
		}

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return "", "", err
		}

		return resultText, imgPath, nil

	}

	return bto.Info.Name, "", nil
}

func (c *Client) AddByFile(operation string) (string, error) {
	if operation == "file+add-no" {
		c.storage.tFile = nil
		return "Okay", nil
	}

	pathKey := strings.Split(operation, "-")[1]
	path := filepath.Join(c.API.Session.DownloadDir + c.Categories[pathKey].Path)

	base64Str := base64.StdEncoding.EncodeToString(c.storage.tFile)

	res, err := c.API.AddTorrent(transmission.AddTorrentArg{
		DownloadDir: path,
		Metainfo:    base64Str,
		Paused:      false,
	})
	if err != nil {
		return "", err
	}

	_ = c.updateCache(context.TODO())
	return fmt.Sprintf("Successfully added\n`%s`\nID:`%d`", res.Name, res.ID), nil
}

func (c *Client) Details(hash string) (Torrent, error) {
	torrent, ok := c.cache.GetByHash(hash)
	if !ok {
		return Torrent{}, errors.New("torrent not found")
	}

	var active bool
	if torrent.Status != transmission.StatusStopped {
		active = true
	}

	icon, status := ParseStatus(torrent.Status)
	var _error bool
	if torrent.ErrorString != "" {
		_error = true
		icon = "🔥️"
	}

	return Torrent{
		ID:             torrent.ID,
		Peers:          len(*torrent.Peers),
		Downloading:    torrent.Status == transmission.StatusDownloading,
		PosInQ:         torrent.QueuePosition,
		Active:         active,
		Error:          _error,
		Name:           torrent.Name,
		Status:         status,
		StatusCode:     torrent.Status,
		Icon:           icon,
		ErrorString:    torrent.ErrorString,
		Size:           glh.ConvertBytes(float64(torrent.TotalSize), glh.Size),
		DownloadedSize: glh.ConvertBytes(float64(torrent.LeftUntilDone), glh.Size),
		Dspeed:         glh.ConvertBytes(float64(torrent.RateDownload), glh.Speed),
		Uspeed:         glh.ConvertBytes(float64(torrent.RateUpload), glh.Speed),
		Percents:       fmt.Sprintf("%.2f%%", torrent.PercentDone*100.0),
	}, nil
}

func (c *Client) Delete(hash string, rmfiles bool) error {
	t, ok := c.cache.GetByHash(hash)
	if !ok {
		return ErrorTorrentNotFound
	}
	err := c.API.RemoveTorrents([]*transmission.Torrent{t}, rmfiles)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetFiles(hash string) []TorrentFilesItem {
	t, _ := c.cache.GetByHash(hash)
	files := *t.Files
	filesStats := *t.FileStats
	var result []TorrentFilesItem

	for i := range len(files) {
		result = append(result, TorrentFilesItem{
			Name:        files[i].Name,
			Size:        glh.ConvertBytes(float64(files[i].Length), glh.Size),
			Downloading: filesStats[i].Wanted,
		})
	}

	return result
}

func (c *Client) StartStop(hash, op string) error {
	t, _ := c.cache.GetByHash(hash)

	switch op {
	case "start":
		if err := t.Start(); err != nil {
			return err
		}
	case "stop":
		if err := t.Stop(); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("torrent state change timeout exceeded")
		case <-ticker.C:
			t, _ := c.cache.GetByHash(hash)
			switch op {
			case "start":
				if t.Status != transmission.StatusStopped {
					return nil
				}
			case "stop":
				if t.Status == transmission.StatusStopped {
					return nil
				}
			}
		}
	}
}

func (c *Client) Priority(hash string, action string) error {
	t, _ := c.cache.GetByHash(hash)

	whatS := strings.Split(action, "-")[1]
	switch whatS {
	case "top":
		err := c.API.QueueMoveTop([]*transmission.Torrent{t})
		if err != nil {
			return err
		}
	case "up":
		err := c.API.QueueMoveUp([]*transmission.Torrent{t})
		if err != nil {
			return err
		}
	case "down":
		err := c.API.QueueMoveDown([]*transmission.Torrent{t})
		if err != nil {
			return err
		}
	case "bottom":
		err := c.API.QueueMoveBottom([]*transmission.Torrent{t})
		if err != nil {
			return err
		}
	case "no":
		return nil
	default:
		return errors.New("nope, failed")
	}

	return nil
}
