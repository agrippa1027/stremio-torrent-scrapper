package stream

import (
	"context"
	"time"

	"scrapper/internal/filter"
	"scrapper/internal/jackett"
	"scrapper/internal/tmdb"
	"scrapper/internal/torbox"

	"github.com/charmbracelet/log"
	"github.com/redis/go-redis/v9"
)

type StreamService struct {
	redis      *redis.Client
	tmdbClient *tmdb.Client
	jackett    *jackett.Client
	torbox     *torbox.Client
	logger     *log.Logger

	// configs
	timeout time.Duration
}

func New(
	tmdbClient *tmdb.Client,
	jackett *jackett.Client,
	torbox *torbox.Client,
	logger *log.Logger,
) (*StreamService, error) {
	return &StreamService{
		jackett: jackett,
		torbox:  torbox,
		logger:  logger,
	}, nil
}

func (s *StreamService) Search(ctx context.Context, query SearchQuery) ([]StremioStream, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	result, err := s.parseTorrents(ctx, query)

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	streams := s.buildTorboxStreamLinks(ctx, result, query.Type, query.Season, query.Episode)

	return streams, nil
}

func (s *StreamService) parseTorrents(ctx context.Context, query SearchQuery) ([]Stream, error) {
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

	streams := make([]Stream, 0, len(results))
	for _, result := range results {
		parsed := filter.ParseTorrent(ctx, result.Title)
		if parsed == nil {
			continue
		}

		if parsed.Title != query.Title {
			s.logger.Warn("title mismatch",
				"parsed", parsed.Title,
				"tmdb", query.Title,
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

	return streams, nil
}
