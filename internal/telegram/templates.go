package telegram

import (
	"strings"
	"text/template"
)

func TmplTorrent() *template.Template {
	lines := []string{
		"{{ .Icon }} | `{{ .Name }}` | ID: {{ .ID }}",
		"Size: {{ .Size }}",
		"Status: {{ .Status }}",
		"Place in queue: {{ .PosInQ }}",
		"`===========================================`",
		"{{ if .Active }}Peers: {{ .Peers }}{{ end }}",
		"{{ if .Active }}⬇️ Downloading: {{ .Dspeed }} | ⬆️ Uploading: {{ .Uspeed }}{{ end }}",
		"{{ if .Downloading }}Percent done: {{ .Percents }}{{ end }}",
		"{{ if .Error }}{{ .ErrorString }}{{ end }}",
	}

	result := template.Must(
		template.New("torrentDetails").
			Parse(strings.Join(lines, "\n")))

	return result
}

func TmplStatus() *template.Template {
	lines := []string{
		"Free Space: `{{ .FreeSpace }}`",
		"Current:",
		"`▶️ Active: {{ .Active }} || ⏸️ Paused: {{ .Paused }}`",
		"`⬇️ Downloading: {{ .DownloadS }} || ⬆️ Uploading: {{ .UploadS }}`",
		"`⬇️ Downloaded: {{ .Downloaded }} || ⬆️ Uploaded: {{ .Uploaded }}`",
	}

	result := template.Must(
		template.New("status").
			Parse(strings.Join(lines, "\n")))

	return result
}

func TmplConfig() *template.Template {
	lines := []string{
		"Default download dir: `{{ .DownloadDir }}`",
		"Start download after add: {{if .StartAdded}} ✔️ {{else}} ❌ {{end}}",
		"{{if .DownloadQEn}}Download Queue size: {{.DownloadQSize}} {{end}}",
		"{{if .SpeedLimitDEn}}Download speed limit:  ✔️ Limit: {{.SpeedLimitD}} {{else}}Download speed limit: ❌ {{end}}",
		"{{if .SpeedLimitUEn}}Upload speed limit:  ✔️ Limit: {{.SpeedLimitU}} {{else}}Upload speed limit: ❌ {{end}}",
	}

	result := template.Must(
		template.New("sessConfig").
			Parse(strings.Join(lines, "\n")))

	return result
}

func TmplTorrentListItem() *template.Template {
	lines := []string{
		"`{{ .Icon }} | {{ .Status }} | ID: {{ .ID }}`",
		"{{ .Name }}",
		"{{ if .ErrorString -}}",
		"`----`",
		"{{ .ErrorString }}",
		"{{ end }}",
	}

	result := template.Must(
		template.New("trListItem").
			Parse(strings.Join(lines, "\n")))

	return result
}

func TmplTorrentFilesListItem() *template.Template {
	lines := []string{
		"`{{ .Name }}`",
		"Size: `{{ .Size }}`",
		"To Download: {{ if .Downloading }} ✔️ {{else}} ❌ {{end}}",
	}

	result := template.Must(
		template.New("trFilesListItem").
			Parse(strings.Join(lines, "\n")))

	return result
}
