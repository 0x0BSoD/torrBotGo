package events

type Type string

const (
	EventTorrentAdded         Type = "torrent.added"
	EventTorrentAddFailed     Type = "torrent.add.failed"
	EventTorrentStarted       Type = "torrent.started"
	EventTorrentStopped       Type = "torrent.stopped"
	EventTorrentRemoved       Type = "torrent.removed"
	EventTorrentRequeued      Type = "torrent.requeued"
	EventTorrentDownloadDone  Type = "torrent.download.done"
	EventTorrentCheckingFiles Type = "torrent.checking.files"
	EventTorrentMagentParsed  Type = "torrent.magnet.parsed"
)

type Event struct {
	Type       Type
	TorrentID  int
	Name       string
	Err        error
	Text       string
	Meta       map[string]any
	MagnetLink string
	MessageID  int
}
