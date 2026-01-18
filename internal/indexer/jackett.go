package indexer

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"scrapper/internal/tmdb"

	jackett "github.com/webtor-io/go-jackett"
)

type Jackett struct {
	client *jackett.Client
	logger *slog.Logger
	tmdb   *tmdb.Client
}

func NewJackett(client *http.Client, apiURL string, apiKey string, tmdbClient *tmdb.Client, logger *slog.Logger) (*Jackett, error) {
	j, err := jackett.New(jackett.Settings{
		ApiURL: apiURL,
		ApiKey: apiKey,
		Client: client,
	})
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &Jackett{client: j, logger: logger, tmdb: tmdbClient}, nil
}

func (j *Jackett) Name() string {
	return "Jackett"
}

func (j *Jackett) Search(ctx context.Context, query SearchQuery) ([]Result, error) {
	// imdbID := normalizeIMDBID(query.ID)
	imdbID := query.ID
	j.logger.Debug("Jackett search",
		"imdb_id", imdbID,
		"query", query.Query,
		"type", query.Type,
		"season", query.Season,
		"episode", query.Episode,
	)
	var results []jackett.Result
	var err error
	switch query.Type {
	case "movie":
		if imdbID != "" {
			req := jackett.NewMovieSearch().WithIMDBID(imdbID).Build()
			results, err = j.client.Fetch(ctx, req)
		} else {
			req := jackett.NewRawSearch().WithQuery(query.Query).Build()
			results, err = j.client.Fetch(ctx, req)
		}
	case "series":
		results, err = j.searchSeries(ctx, query, imdbID)
	default:
		req := jackett.NewRawSearch().WithQuery(query.Query).Build()
		results, err = j.client.Fetch(ctx, req)
	}
	if err != nil {
		j.logger.Error("Jackett fetch error", "error", err)
		return nil, err
	}

	out := make([]Result, 0, len(results))
	for _, result := range results {
		out = append(out, Result{
			Title:    result.Title,
			InfoHash: result.InfoHash,
			Magnet:   result.MagnetURI,
			Size:     int64(result.Size),
			Seeds:    int(result.Seeders),
			Source:   result.Tracker,
			PageURL:  result.Link,
		})
	}

	return out, nil
}

var imdbPattern = regexp.MustCompile(`^tt\\d{7,}$`)

func normalizeIMDBID(id string) string {
	id = strings.TrimSpace(id)
	if imdbPattern.MatchString(id) {
		return id
	}
	return ""
}

func isUnsupportedIMDBError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "imdbid is not supported for tv search")
}

func (j *Jackett) tmdbTitleByIMDB(ctx context.Context, imdbID string) (string, error) {
	if j.tmdb == nil {
		return "", tmdb.ErrMissingAPIKey
	}
	return j.tmdb.TVTitleByIMDB(ctx, imdbID)
}

func (j *Jackett) searchSeries(ctx context.Context, query SearchQuery, imdbID string) ([]jackett.Result, error) {
	title := ""
	if t, err := j.tmdbTitleByIMDB(ctx, imdbID); err == nil {
		title = t
	}

	// requests are ordered by title whole season, title specific episode, IMDB whole season, IMDB specific episode
	requests := j.seriesRequests(query, imdbID, title)
	var lastErr error

	for _, req := range requests {
		results, err := j.client.Fetch(ctx, req)
		if err != nil {
			if imdbID != "" && isUnsupportedIMDBError(err) {
				// This means there in no point in going back. as we have tried
				// almost every to way to search
				j.logger.Warn("Jacket threw and isUnsupportedIMDBError", "err", err)
				lastErr = err
				break
			}
			return nil, err
		}
		if len(results) > 0 {
			return results, nil
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, nil
}

func (j *Jackett) seriesRequests(query SearchQuery, imdbID string, title string) []*jackett.FetchRequest {
	var requests []*jackett.FetchRequest

	/*
	*  Most indexers do not support TV search by IMDB ID, so we try multiple methods
	*  so we are first trying with title then moving onto IMDB id
	*  it is also preferable to get the complete season then get the individual episode from then
	*  complete season torrent
	 */
	if title != "" && query.Season > 0 {
		requests = append(requests, j.buildTVSearchByTitle(title, query.Season, 0, true))
		requests = append(requests, j.buildTVSearchByTitle(title, query.Season, 0, false))
	}

	if title != "" && query.Season > 0 && query.Episode > 0 {
		requests = append(requests, j.buildTVSearchByTitle(title, query.Season, query.Episode, false))
	}

	if imdbID != "" {
		requests = append(requests, j.buildTVSearchByIMDB(imdbID, query.Season, 0))
		if query.Season > 0 && query.Episode > 0 {
			requests = append(requests, j.buildTVSearchByIMDB(imdbID, query.Season, query.Episode))
		}
	}

	if query.Query != "" {
		requests = append(requests, jackett.NewRawSearch().WithQuery(query.Query).Build())
	}

	return requests
}

func (j *Jackett) buildTVSearchByIMDB(imdbID string, season int, episode int) *jackett.FetchRequest {
	tv := jackett.NewTVSearch().WithIMDBID(imdbID)
	if season > 0 {
		tv = tv.WithSeason(uint(season))
	}
	if episode > 0 {
		tv = tv.WithEpisode(uint(episode))
	}
	return tv.Build()
}

func (j *Jackett) buildTVSearchByTitle(title string, season int, episode int, multiSeason bool) *jackett.FetchRequest {
	query := title
	if season > 0 {
		if multiSeason {
			query = strings.TrimSpace(query + " S01-S" + twoDigits(season))
		} else if episode > 0 {
			query = strings.TrimSpace(query + " S" + twoDigits(season) + "E" + twoDigits(episode))
		} else {
			query = strings.TrimSpace(query + " S" + twoDigits(season))
		}
	}
	return jackett.NewTVSearch().WithQuery(query).Build()
}

func twoDigits(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}
