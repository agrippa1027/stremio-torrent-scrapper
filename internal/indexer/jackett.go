package indexer

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
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
	var req *jackett.FetchRequest
	if imdbID != "" {
		switch query.Type {
		case "movie":
			req = jackett.NewMovieSearch().
				WithIMDBID(imdbID).
				Build()
			j.logger.Debug("Jackett movie search",
				"imdb_id", imdbID,
			)
		case "series":
			j.logger.Debug("Jackett tv search",
				"imdb_id", imdbID,
				"season", query.Season,
				"episode", query.Episode,
			)
			tv := jackett.NewTVSearch().
				WithIMDBID(imdbID)
			if query.Season > 0 {
				tv = tv.WithSeason(uint(query.Season))
			}
			if query.Episode > 0 {
				tv = tv.WithEpisode(uint(query.Episode))
			}
			j.logger.Debug("Jackett TV search built",
				"imdb_id", imdbID,
				"season", query.Season,
				"episode", query.Episode,
			)
			req = tv.Build()
		}
	}

	results, err := j.client.Fetch(ctx, req)
	if err != nil {
		j.logger.Warn("Jackett fetch failed", "error", err)
		if query.Type == "series" && imdbID != "" && isUnsupportedIMDBError(err) {
			title, titleErr := j.tmdbTitleByIMDB(ctx, imdbID)
			if titleErr != nil {
				j.logger.Warn("tmdb imdb lookup failed",
					"imdb_id", imdbID,
					"error", titleErr,
				)
				return nil, err
			}
			if title == "" {
				return nil, err
			}
			tv := jackett.NewTVSearch().WithQuery(title)
			if query.Season > 0 {
				tv = tv.WithSeason(uint(query.Season))
			}
			if query.Episode > 0 {
				tv = tv.WithEpisode(uint(query.Episode))
			}
			fallbackReq := tv.Build()
			j.logger.Debug("Jackett tv fallback search",
				"imdb_id", imdbID,
				"title", title,
				"season", query.Season,
				"episode", query.Episode,
			)
			results, err = j.client.Fetch(ctx, fallbackReq)
			if err != nil {
				return nil, err
			}
		} else {
			j.logger.Error("Jackett fetch error", "error", err)
			return nil, err
		}
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
