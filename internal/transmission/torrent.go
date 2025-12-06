package transmission

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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

	c.Storage.tFile, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(bytes.NewReader(c.Storage.tFile), &bto)
	if err != nil {
		return "", "", err
	}

	// torrent from rutracker
	if bto.Comment != "" {
		imgURL, err := getImgURLRutracker(bto.Comment)
		if err != nil {
			return bto.Info.Name, "", nil
		}

		_, err = url.ParseRequestURI(imgURL)
		if err != nil {
			return bto.Info.Name, "", nil
		}

		client := httpClient()
		resp, err := client.Get(imgURL)
		if err != nil {
			return bto.Info.Name, "", nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return bto.Info.Name, "", nil
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

		return bto.Info.Name, imgPath, nil

	}

	return bto.Info.Name, "", nil
}
