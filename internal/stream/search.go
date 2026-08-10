package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"scrapper/internal/filter"
	"scrapper/internal/indexer"
	"scrapper/internal/tmdb"
	"scrapper/logging"
	"scrapper/redis"
)

func SearchStreams(
	ctx context.Context,
	query SearchQuery,
	indexers []indexer.Indexer,
	tmdbClient *tmdb.Client,
) []Stream {
	logger := logging.FromContext(ctx)

	logger.Debug("stream request",
		"type", query.Type,
		"id", query.ID,
		"season", query.Season,
		"episode", query.Episode,
	)

	// TODO: Get timeout from config
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// TODO: wiered ass api. have to deal with this later
	tmdbResult, err := tmdbClient.TVTitleByIMDB(ctx, query.ID)

	data, err := json.Marshal(tmdbResult)

	if err != nil {
	    logger.Error("json.Marshal", "err", err)
	}

	redisClient := redis.NewRedisClient()
	res, err := redisClient.Set(ctx, fmt.Sprintf("tmdb:%s", query.ID), data, 0).Result()
	if err != nil {
		logger.Error("redisClient.HSet", "err", err)
	}
	logger.Debug("redisClient.HSet", "res", res)

	if err != nil {
		logger.Error("tmdbClient.TVTitleByIMDB", "err", err)
		return nil
	}
	tmdbTitle := tmdbResult[0].Name

	type response struct {
		results   []indexer.Torrent
		err       error
		name      string
		tmdbTitle string
	}

	// NOTE: currently the go routines and buffer for the channel doesn't matter
	// we only have one indexeer (jackett) and it only produce a list of torretns
	// from the api. Thus we are unabel to produce a goroutine to handle the torrents
	// from every tracker in parallel.
	ch := make(chan response, len(indexers))

	var wg sync.WaitGroup
	wg.Add(len(indexers))

	for _, idx := range indexers {
		go func() {
			defer wg.Done()
			results, err := searchIndexer(ctx, indexer.SearchQuery{
				Type:      indexer.QType(query.Type),
				TMDBTitle: tmdbTitle,
				ID:        query.ID,
				Season:    query.Season,
				Episode:   query.Episode,
			}, idx)
			ch <- response{
				results:   results,
				err:       err,
				name:      idx.Name(),
				tmdbTitle: tmdbTitle,
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	combined := make([]indexer.Torrent, 0, 20)

	for v := range ch {
		if v.err != nil {
			logger.Warn("indexer search failed",
				"indexer", v.name,
				"id", query.ID,
				"error", v.err,
			)
			continue
		}
		if len(v.results) == 0 {
			logger.Info("indexer search empty",
				"indexer", v.name,
				"id", query.ID,
			)
			continue
		}
		combined = append(combined, v.results...)
		tmdbTitle = v.tmdbTitle
	}

	paresedTorrents := parseTorrents(ctx, combined, tmdbTitle)

	return paresedTorrents
}

func searchIndexer(ctx context.Context, query indexer.SearchQuery,
	idx indexer.Indexer) ([]indexer.Torrent, error) {
	results, err := idx.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func parseTorrents(ctx context.Context, results []indexer.Torrent,
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
