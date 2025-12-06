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

	"github.com/0x0BSoD/transmission"
)

type filesList struct {
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

var ErrorFilterNotFound = errors.New("unknown filter")

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

func (c *Client) AddTorrentByFileDialog(directURL string) (string, string, error) {
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

func (c *Client) AddTorrentByFile(operation string) (string, error) {
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
