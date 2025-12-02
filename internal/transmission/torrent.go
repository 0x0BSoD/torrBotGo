package transmission

import "github.com/0x0BSoD/transmission"

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
	all showFilter = iota
	active
	notActive
)

type filesList struct {
	Name        string
	Size        string
	Downloading bool
}

func (c *Client) Torrents(sf showFilter) (map[string]*transmission.Torrent, error) {
	items, _ := c.cache.Snapshot()

	result := make(map[string]*transmission.Torrent)

	for _, torrent := range items {
		switch sf {
		case all:
			result[torrent.HashString] = torrent
		case active:
			if torrent.Status != transmission.StatusStopped && torrent.ErrorString == "" {
				result[torrent.HashString] = torrent
			}
		case notActive:
			if torrent.Status == transmission.StatusStopped {
				result[torrent.HashString] = torrent
			}
		}
	}

	return result, nil
}
