package stream

import (
	"context"
	"time"

	"scrapper/internal/filter"
	"scrapper/internal/jackett"
	"scrapper/internal/tmdb"
	"scrapper/logging"

	"github.com/charmbracelet/log"
	"github.com/redis/go-redis/v9"
)

type SearchService struct {
	redis      *redis.Client
	tmdbClient *tmdb.Client
	jackett    *jackett.Jackett
	logger     *log.Logger

	// configs
	timeout time.Duration
}

func (s *SearchService) Search(ctx context.Context, query SearchQuery) ([]Stream, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	results, err := s.jackett.Search(ctx, jackett.SearchQuery{
		Type:      jackett.QType(query.Type),
		TMDBTitle: query.Title,
		ID:        query.ID,
		Season:    query.Season,
		Episode:   query.Episode,
	})

	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}

	paresedTorrents := parseTorrents(ctx, results, query.Title)

	return paresedTorrents, nil
}

func parseTorrents(ctx context.Context, results []jackett.Torrent,
	tmdbTitle string) []Stream {
	logger := logging.FromContext(ctx)
	streams := make([]Stream, 0, len(results))
	for _, result := range results {

		parsed := filter.ParseTorrent(ctx, result.Title)

		if parsed == nil {
			continue
		}

		if parsed.Title != tmdbTitle {
			logger.Warn("title mismatch",
				"parsed", parsed.Title,
				"tmdb", tmdbTitle,
			)
			continue
		}

		stream := Stream{
			Title:      parsed.Title,
			Year:       parsed.Year,
			Seasons:    parsed.Seasons,
			Episodes:   parsed.Episodes,
			Complete:   parsed.Complete,
			Uncensored: parsed.Uncensored,
			Unrated:    parsed.Unrated,
			Resolution: parsed.Resolution,
			Codec:      parsed.Codec,
			HDR:        parsed.HDR,

			InfoHash: result.InfoHash,
			Magnet:   result.Magnet,
			Size:     result.Size,
			Tracker:  result.Tracker,
			Files:    uint(result.Files),
		}

		streams = append(streams, stream)
	}
	return streams
}
