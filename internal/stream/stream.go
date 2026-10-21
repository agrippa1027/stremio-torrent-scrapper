package stream

// TODO:
//
// do we really need this spereate file for one method

import (
	"context"
	"fmt"
	"net/url"
)

func (s *StreamService) buildTorboxStreamLinks(
	ctx context.Context,
	results []Stream,
	typ string,
	season int,
	episode int) []StremioStream {

	streams := make([]StremioStream, 0, len(results))

	for _, result := range results {
		infoHash := result.InfoHash
		if infoHash == "" {
			s.logger.Warn("[Stream][buildTorboxStreamLinks] infoHash is empty, skipping",
				"result", result,
			)
			continue
		}

		cached, err := s.torbox.CheckCached(ctx, infoHash)
		if err != nil {
			s.logger.Warn("[Stream][buildTorboxStreamLinks] torbox checkcached failed",
				"hash", infoHash,
				"error", err,
			)
			continue
		}

		label := "Torbox [add]"
		link := fmt.Sprintf("%s/torbox/stream/%s?cached=0&title=%s",
			s.torbox.BaseURL,
			infoHash,
			url.QueryEscape(result.Title),
		)
		if typ == "series" && season > 0 && episode > 0 {
			link += fmt.Sprintf("&season=%d&episode=%d", season, episode)
		}

		if cached {
			label = "Torbox [cached]"
			directURL, err := s.torbox.TorboxStreamURL(ctx, infoHash, result.Title, true, season, episode, result.Magnet)
			if err != nil {
			} else {
				link = directURL
			}
		}

		streams = append(streams, StremioStream{
			Title:        result.Title + " • " + label,
			URL:          link,
			InfoHash:     infoHash,
			TorrentTitle: result.TorrentTitle,
		})
	}

	return streams
}
