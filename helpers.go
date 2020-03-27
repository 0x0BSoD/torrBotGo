package main

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/0x0BSoD/telegram-bot-api"
	"github.com/PuerkitoBio/goquery"
)

func sendError(text string) {
	if ctx.chatID == 0 {
		fmt.Errorf("chatID empty, %s", text)
		return
	}

	msg := tgbotapi.NewVideoUpload(ctx.chatID, "error.mp4")
	msg.Caption = text
	if _, err := ctx.Bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func parseStatus(s int) (string, string) {
	var icon string
	var status string

	switch s {
	case 0:
		icon = "⏹️️"
		status = "Stopped"
	case 1:
		icon = "▶️️"
		status = "Queued to check files"
	case 2:
		icon = "▶️"
		status = "Checking files"
	case 3:
		icon = "▶️️"
		status = "Queued to download"
	case 4:
		icon = "▶️"
		status = "Downloading"
	case 5:
		icon = "▶️️"
		status = "'Queued to seed"
	default:
		icon = "▶️️"
		status = "Seeding"
	}

	return icon, status
}

func sendNewMessage(chatID int64, text string, replyMarkup *tgbotapi.InlineKeyboardMarkup) error {
	if ctx.chatID == 0 {
		return errors.New("chatID empty")
	}

	if text == "" {
		return fmt.Errorf("message cannot be empty")
	}

	msg := tgbotapi.NewMessage(chatID, escapeAll(text))
	msg.ParseMode = "MarkdownV2"

	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup
	}

	if _, err := ctx.Bot.Send(msg); err != nil {
		return err
	}

	return nil
}

func sendEditedMessage(chatID int64, messageID int, text string, replyMarkup *tgbotapi.InlineKeyboardMarkup) error {
	if text == "" {
		return fmt.Errorf("message cannot be empty")
	}

	msg := tgbotapi.NewEditMessageText(chatID, messageID, escapeAll(text))
	msg.ParseMode = "MarkdownV2"

	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup
	}

	if _, err := ctx.Bot.Send(msg); err != nil {
		return err
	}

	return nil
}

func sendNewImagedMessage(chatID int64, text string, image io.Reader, replyMarkup *tgbotapi.InlineKeyboardMarkup) error {

	hasher := sha1.New()
	tmHash := strconv.Itoa(time.Now().Nanosecond())
	hasher.Write([]byte(tmHash))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	file, err := os.Create("/tmp/" + sha)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, image)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewPhotoUpload(chatID, "/tmp/"+sha)
	msg.ParseMode = "MarkdownV2"

	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup
	}

	msg.Caption = escapeAll(text)
	if _, err := ctx.Bot.Send(msg); err != nil {
		return err
	}

	return nil
}

// rutracker
func getImgFromTrackerRutracker(url string) (string, error) {

	if !strings.HasPrefix(url, "https://rutracker.org/") {
		return "", errors.New("not a rutracker")
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	var imgURL string
	doc.Find(".postImgAligned").Each(func(i int, s *goquery.Selection) {
		imgURL, _ = s.Attr("title")
	})

	return imgURL, nil
}

func httpClient() *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	return &client
}

func escapeAll(text string) string {
	re := strings.NewReplacer("-", "\\-",
		".", "\\.",
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"|", "\\|",
		"!", "\\!")
	return re.Replace(text)
}
