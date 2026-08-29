package jackett

import (
	"context"
	"scrapper/logging"

	jackett "github.com/webtor-io/go-jackett"
)

func (j *Jackett) searchSeries(ctx context.Context, query SearchQuery, imdbID string) ([]jackett.Result, error) {
	logger := logging.FromContext(ctx)

	logger.Debug("searchSeries", "title", query.TMDBTitle, "season", query.Season, "episode", query.Episode)

	// requests are ordered by title whole season, title specific episode,
	// IMDB whole season, IMDB specific episode
	//
	// We have complete season
	// specific episode
	//
	// TODO: how are we going to give stream link for a complete season
	// is it supported by default in stremio
	// requests := j.seriesRequests(query, imdbID, title)
	//
	request := j.buildTVSearchByTitle(query.TMDBTitle, query.Season, query.Episode)

	results, err := j.client.Fetch(ctx, request)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// TODO: We are building the TV search by title query manually.
// should consider using the Jackett API to search by title/season/episode fields instead
func (j *Jackett) buildTVSearchByTitle(title string, season, episode int) *jackett.FetchRequest {
	return jackett.NewTVSearch().
		WithQuery(title).
		WithSeason(uint(season)).
		WithEpisode(uint(episode)).
		Build()
}
