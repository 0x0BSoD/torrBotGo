package transmission

import (
	"errors"

	"github.com/0x0BSoD/transmission"
)

// type Torrent struct {
// 	ID             int
// 	Peers          int
// 	Downloading    bool
// 	Active         bool
// 	Name           string
// 	Status         string
// 	Icon           string
// 	Error          bool
// 	ErrorString    string
// 	DownloadedSize string
// 	Size           string
// 	Comment        string
// 	Hash           string
// 	PosInQ         int
// 	Dspeed         string
// 	Uspeed         string
// 	Percents       string
// }

type filesList struct {
	Name        string
	Size        string
	Downloading bool
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
