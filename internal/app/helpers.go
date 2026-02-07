// Package app provides the core application logic for torrBotGo.
// It handles Telegram message routing, command processing, and user interactions.
package app

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	glh "github.com/0x0BSoD/goLittleHelpers"
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

func sendMessageWrapper(tClient *telegram.Client, tmpl *template.Template, kbd, data any) {
	toSend := data

	if tmpl != nil {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			tClient.SendError(fmt.Sprintf("template execute failed, %v", err))
			return
		}
		toSend = buf.String()
	}

	if err := tClient.SendMessage(toSend.(string), kbd); err != nil {
		tClient.SendError(fmt.Sprintf("send failed, %v", err))
		return
	}
}

func editMessageWrapperHash(messageID int, oldMessage string, tClient *telegram.Client, tmpl *template.Template, kbd, data any) {
	toSend := data

	if tmpl != nil {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			tClient.SendError(fmt.Sprintf("template execute failed, %v", err))
			return
		}
		toSend = buf.String()
	}

	oldHash := glh.GetMD5Hash(strings.ReplaceAll(strings.ReplaceAll(oldMessage, "`", ""), "\n", ""))
	newHash := glh.GetMD5Hash(strings.ReplaceAll(strings.ReplaceAll(toSend.(string), "`", ""), "\n", ""))
	if newHash == oldHash {
		return
	}

	if err := tClient.SendEditedMessage(messageID, toSend.(string), kbd); err != nil {
		tClient.SendError(fmt.Sprintf("send torrent deleted failed, %v", err))
		return
	}
}

func addTorrent(query string, messageID int, tClient *telegram.Client, trClient *intTransmission.Client) {
	var (
		text string
		err  error
	)

	if strings.HasPrefix(query, "file+add-") {
		text, err = trClient.AddByFile(query)
	} else {
		text, err = trClient.AddByMagnet(query)
	}
	if err != nil {
		tClient.SendError(fmt.Sprintf("add torrent failed, %v", err))
		return
	}

	if messageID != -1 {
		if err := tClient.RemoveMessage(messageID); err != nil {
			tClient.SendError(fmt.Sprintf("remove message failed, %v", err))
			return
		}
	}

	if err := tClient.SendMessage(text, nil); err != nil {
		tClient.SendError(fmt.Sprintf("send message failed, %v", err))
		return
	}
}
