package indexer

import (
	"context"
	"fmt"
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
		req := jackett.NewMovieSearch().WithIMDBID(imdbID).Build()
		results, err = j.client.Fetch(ctx, req)

	case "series":
		results, err = j.searchSeries(ctx, query, imdbID)

	default:
		return nil, fmt.Errorf("unsupported search type: %s", query.Type)
	}

	if err != nil {
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

	var results []jackett.Result
	for _, req := range requests {
		j.logger.Debug("Jackett series search request", "request", req.TV)
		fetch, err := j.client.Fetch(ctx, req)
		if err != nil {
			if imdbID != "" && isUnsupportedIMDBError(err) {
				// This means there in no point in going back. as we have tried
				// almost every to way to search
				j.logger.Warn("Jacket threw and isUnsupportedIMDBError", "err", err)
				break
			}
		}
		if len(fetch) > 0 {
			filtered := filterSeriesResults(fetch, title)
			j.logger.Debug("Jackett series search got results",
				"numResults", len(fetch),
				"filtered", len(filtered),
			)
			results = append(results, filtered...)
		}
	}

	return results, nil
}

func filterSeriesResults(results []jackett.Result, title string) []jackett.Result {
	if title == "" {
		return results
	}

	normalizedTitle := normalizeSeriesTitle(title)
	if normalizedTitle == "" {
		return results
	}

	filtered := make([]jackett.Result, 0, len(results))
	for _, result := range results {
		if matchesSeriesTitle(result.Title, normalizedTitle) {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

func matchesSeriesTitle(rawTitle string, normalizedTitle string) bool {
	candidate := normalizeSeriesTitle(rawTitle)
	if candidate == "" {
		return false
	}

	titleTokens := strings.Fields(normalizedTitle)
	candidateTokens := strings.Fields(candidate)
	if len(titleTokens) == 0 || len(candidateTokens) < len(titleTokens) {
		return false
	}

	for i := 0; i+len(titleTokens) <= len(candidateTokens); i++ {
		if !tokensMatch(candidateTokens[i:i+len(titleTokens)], titleTokens) {
			continue
		}
		nextIndex := i + len(titleTokens)
		if nextIndex >= len(candidateTokens) {
			return true
		}
		nextToken := candidateTokens[nextIndex]
		if isAllowedSeriesSuffix(nextToken) || isSeasonOrEpisodeToken(nextToken) {
			return true
		}
	}

	return false
}

func tokensMatch(candidate []string, title []string) bool {
	for i := range title {
		if candidate[i] != title[i] {
			return false
		}
	}
	return true
}

var (
	seasonEpisodeTokenPattern = regexp.MustCompile(`(?i)^s?\d{1,2}e\d{1,2}$|^\d{1,2}x\d{1,2}$`)
	seasonTokenPattern        = regexp.MustCompile(`(?i)^s\d{1,2}$|^season\d{1,2}$|^s\d{1,2}-s\d{1,2}$|^s\d{1,2}to\d{1,2}$`)
)

func isSeasonOrEpisodeToken(token string) bool {
	return seasonEpisodeTokenPattern.MatchString(token) || seasonTokenPattern.MatchString(token)
}

func isAllowedSeriesSuffix(token string) bool {
	switch token {
	case "uk", "us", "u", "u.s", "u.s.", "u.k", "u.k.", "complete", "collection", "series", "season", "seasons":
		return true
	default:
		return false
	}
}

func normalizeSeriesTitle(raw string) string {
	normalized := strings.ToLower(raw)
	replacer := strings.NewReplacer(
		".", " ",
		"_", " ",
		"-", " ",
		":", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"'", " ",
		"\"", " ",
	)
	normalized = replacer.Replace(normalized)
	fields := strings.Fields(normalized)
	return strings.Join(fields, " ")
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
