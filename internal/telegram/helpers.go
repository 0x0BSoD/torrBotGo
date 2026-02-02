// Package telegram provides Telegram Bot API integration for torrBotGo.
// It handles message sending, inline keyboards, and communication with Telegram servers.
package telegram

import "strings"

func escapeAll(text string) string {
	re := strings.NewReplacer("-", "\\-",
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return re.Replace(text)
}
