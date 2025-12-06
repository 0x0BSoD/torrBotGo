package app

import (
	"bytes"

	"github.com/0x0BSoD/transmission"

	"github.com/0x0BSoD/torrBotGo/internal/telegram"
)

type TorrentStatus struct {
	Icon  string
	Label string
}

var statusMap = map[int]TorrentStatus{
	0: {"⏹️", "Stopped"},
	1: {"▶️", "Queued to check files"},
	2: {"▶️", "Checking files"},
	3: {"▶️", "Queued to download"},
	4: {"▶️", "Downloading"},
	5: {"▶️", "Queued to seed"},
	6: {"▶️", "Seeding"},
}

func parseStatus(code int) (string, string) {
	s, ok := statusMap[code]
	if !ok {
		return "♾️", "Undef"
	}
	return s.Icon, s.Label
}

type input struct {
	ID          int64
	Name        string
	Status      string
	Icon        string
	ErrorString string
}

type category struct {
	Path    string
	Matcher string
}

func renderTorrent(torrent *transmission.Torrent) (string, error) {
	icon, status := parseStatus(torrent.Status)

	var buf bytes.Buffer
	if err := telegram.TmplTorrentListItem().Execute(&buf, input{
		ID:          int64(torrent.ID),
		Name:        torrent.Name,
		Icon:        icon,
		Status:      status,
		ErrorString: torrent.ErrorString,
	}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func extractKeys(input map[string]struct {
	Path    string `yaml:"path"`
	Matcher string `yaml:"matcher"`
},
) []string {
	result := make([]string, len(input))
	i := 0
	for name := range input {
		result[i] = name
		i++
	}
	return result
}
