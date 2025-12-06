package events

type Type string

const (
	EventTorrentDownloadDone Type = "torrent.download.done"
)

type Event struct {
	Type Type
	Text string
}
