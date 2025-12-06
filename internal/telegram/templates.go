package telegram

import (
	"strings"
	"text/template"
)

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
