package app

import (
	"bytes"

	"github.com/0x0BSoD/transmission"

	"github.com/0x0BSoD/torrBotGo/internal/telegram"
	intTransmission "github.com/0x0BSoD/torrBotGo/internal/transmission"
)

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
	icon, status := intTransmission.ParseStatus(torrent.Status)

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
