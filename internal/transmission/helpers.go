// Package transmission provides Transmission RPC client integration for torrBotGo.
// It handles all torrent-related operations including adding, removing, starting,
// stopping torrents, and monitoring torrent status.
package transmission

import (
	"net/http"
	"slices"
	"strings"
)

func httpClient() *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	return &client
}

func matchCategory(input []string, categories map[string]struct {
	Path    string `yaml:"path"`
	Matcher string `yaml:"matcher"`
},
) string {
	for i, j := range categories {
		matchers := strings.Split(j.Matcher, ",")
		for _, k := range input {
			if slices.Contains(matchers, k) {
				return i
			}
		}
	}
	return "noop"
}

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

func ParseStatus(code int) (string, string) {
	s, ok := statusMap[code]
	if !ok {
		return "♾️", "Undef"
	}
	return s.Icon, s.Label
}
