// Package transmission provides Transmission RPC client integration for torrBotGo.
// It handles all torrent-related operations including adding, removing, starting,
// stopping torrents, and monitoring torrent status.
package transmission

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

func fetchPage(url string) (*goquery.Document, error) {
	client := httpClient()
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	utf8Reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(utf8Reader)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func getImgURLRutracker(doc *goquery.Document) string {
	var imgURL string
	doc.Find(".postImgAligned").Each(func(i int, s *goquery.Selection) {
		imgURL, _ = s.Attr("title")
	})

	return imgURL
}

func getCategoryRutracker(doc *goquery.Document) []string {
	var category []string
	doc.Find(".t-breadcrumb-top").Each(func(i int, s *goquery.Selection) {
		aItems := s.Find("a")
		aItems.Each((func(i int, s *goquery.Selection) {
			category = append(category, s.Text())
		}))
	})

	return category
}
