package transmission

import (
	"errors"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

func getImgURLRutracker(url string) (string, error) {
	if !strings.HasPrefix(url, "https://rutracker.org/") {
		return "", errors.New("not a RuTracker")
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
