// Package events provides an event bus system for torrBotGo.
// It enables publish-subscribe communication between components,
// allowing decoupled event handling for system notifications.
package events

type Type string

const (
	EventTorrentDownloadDone Type = "torrent.download.done"
)

type Event struct {
	Type Type
	Text string
}
