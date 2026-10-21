package stream

import (
	"context"
	"fmt"

	"scrapper/internal/filter"
	"scrapper/internal/jackett"
	"scrapper/internal/tmdb"
	"scrapper/internal/torbox"

	"charm.land/log/v2"
)

type StreamService struct {
	tmdbClient *tmdb.Client
	jackett    *jackett.Client
	torbox     *torbox.Client
	logger     *log.Logger
}

func New(
	tmdbClient *tmdb.Client,
	jackett *jackett.Client,
	torbox *torbox.Client,
	logger *log.Logger,
) (*StreamService, error) {
	return &StreamService{
		tmdbClient: tmdbClient,
		jackett:    jackett,
		torbox:     torbox,
		logger:     logger,
	}, nil
}

func (s *StreamService) Search(
	ctx context.Context,
	typeStr,
	idStr string,
	season,
	episode int,
) ([]StremioStream, error) {
	title, err := s.tmdbClient.TitleByIMDB(ctx, idStr)
	if err != nil {
		return nil, fmt.Errorf("[Stream][Search] failed to get title by imdb: %w", err)
	}

	results, err := s.jackett.Search(ctx, jackett.SearchQuery{
		Type:      jackett.QType(typeStr),
		TMDBTitle: title,
		ID:        idStr,
		Season:    season,
		Episode:   episode,
	})

	if err != nil {
		return nil, fmt.Errorf("[Stream][Search] failed to search: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("[Stream][Search] no results found")
	}

	streams := make([]Stream, 0, len(results))
	for _, result := range results {
		parsed := filter.ParseTorrent(ctx, result.Title)
		if parsed == nil {
			continue
		}

		if parsed.Title != title {
			s.logger.Warn("[Stream][Search] title mismatch",
				"parsed", parsed.Title,
				"tmdb", title,
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

			TorrentTitle: result.Title,
		}

		streams = append(streams, stream)
	}

	// TODO:
	// Cache
	stremioStreams := s.buildTorboxStreamLinks(ctx, streams, typeStr, season, episode)

	return stremioStreams, nil
}
