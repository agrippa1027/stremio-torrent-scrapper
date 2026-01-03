package indexer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type X1337 struct {
	client *http.Client
}

func NewX1337(client *http.Client) *X1337 {
	return &X1337{client: client}
}

func (x *X1337) Name() string {
	return "1337x"
}

func (x *X1337) Search(ctx context.Context, query string) ([]Result, error) {
	searchURL := fmt.Sprintf("https://1337x.to/search/%s/1/", url.PathEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	WithUserAgent(req)

	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("1337x status: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	results := make([]Result, 0, 10)
	doc.Find("table.table-list tbody tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i >= 5 {
			return false
		}

		nameSel := s.Find("td.name a").Last()
		title := strings.TrimSpace(nameSel.Text())
		pageHref, _ := nameSel.Attr("href")
		if title == "" || pageHref == "" {
			return true
		}

		seedText := strings.TrimSpace(s.Find("td.seeds").Text())
		seeds, _ := strconv.Atoi(seedText)

		sizeText := strings.TrimSpace(s.Find("td.size").Text())
		size := parseSize(sizeText)

		pageURL := "https://1337x.to" + pageHref
		magnet, infoHash := x.fetchMagnet(ctx, pageURL)

		results = append(results, Result{
			Title:    title,
			Magnet:   magnet,
			InfoHash: infoHash,
			Seeds:    seeds,
			Size:     size,
			Source:   x.Name(),
			PageURL:  pageURL,
		})

		return true
	})

	return results, nil
}

func (x *X1337) fetchMagnet(ctx context.Context, pageURL string) (string, string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", ""
	}
	WithUserAgent(req)

	resp, err := x.client.Do(req)
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ""
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", ""
	}

	magnet := ""
	doc.Find("a").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		href, _ := s.Attr("href")
		if strings.HasPrefix(href, "magnet:?") {
			magnet = href
			return false
		}
		return true
	})

	if magnet == "" {
		return "", ""
	}

	infoHash := extractInfoHash(magnet)
	return magnet, infoHash
}

func extractInfoHash(magnet string) string {
	parts := strings.Split(magnet, "xt=urn:btih:")
	if len(parts) < 2 {
		return ""
	}
	end := strings.IndexAny(parts[1], "&")
	if end == -1 {
		return parts[1]
	}
	return parts[1][:end]
}

func parseSize(raw string) int64 {
	parts := strings.Fields(raw)
	if len(parts) < 2 {
		return 0
	}

	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}

	switch strings.ToUpper(parts[1]) {
	case "KB":
		return int64(value * 1024)
	case "MB":
		return int64(value * 1024 * 1024)
	case "GB":
		return int64(value * 1024 * 1024 * 1024)
	case "TB":
		return int64(value * 1024 * 1024 * 1024 * 1024)
	default:
		return int64(value)
	}
}
