package telegram

import "strings"

func TmplConfig() string {
	lines := []string{
		"{{- /*gotype: github.com/0x0BSoD/transmissionTG.sessConfig*/ -}}",
		"Default download dir: `{{ .DownloadDir }}`",
		"Start download after add: {{if .StartAdded}} ✔️ {{else}} ❌ {{end}}",
		"{{if .DownloadQEn}}Download Queue size: {{.DownloadQSize}} {{end}}",
		"{{if .SpeedLimitDEn}}Download speed limit:  ✔️ Limit: {{.SpeedLimitD}} {{else}}Download speed limit: ❌ {{end}}",
		"{{if .SpeedLimitUEn}}Upload speed limit:  ✔️ Limit: {{.SpeedLimitU}} {{else}}Upload speed limit: ❌ {{end}}",
	}

	return strings.Join(lines, "\n")
}
