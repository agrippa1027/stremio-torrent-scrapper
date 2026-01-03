package addon

import (
	"context"
	"fmt"
	"strings"
	"time"

	"scrapper/internal/indexer"
)

func (a *Addon) searchStreams(ctx context.Context, query string) []indexer.Result {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	type response struct {
		results []indexer.Result
		err     error
		name    string
	}

	ch := make(chan response, len(a.indexers))
	for _, idx := range a.indexers {
		idx := idx
		go func() {
			results, err := idx.Search(ctx, query)
			if err == nil {
				for i := range results {
					if results[i].Source == "" {
						results[i].Source = idx.Name()
					}
				}
			}
			ch <- response{results: results, err: err, name: idx.Name()}
		}()
	}

	combined := make([]indexer.Result, 0, 20)
	for i := 0; i < len(a.indexers); i++ {
		select {
		case res := <-ch:
			if res.err != nil {
				a.logger.Warn("indexer search failed",
					"indexer", res.name,
					"query", query,
					"error", res.err,
				)
				continue
			}
			if len(res.results) == 0 {
				a.logger.Info("indexer search empty",
					"indexer", res.name,
					"query", query,
				)
				continue
			}
			a.logger.Info("indexer search success",
				"indexer", res.name,
				"query", query,
				"results", len(res.results),
			)
			combined = append(combined, res.results...)
		case <-ctx.Done():
			a.logger.Warn("indexer search timeout",
				"query", query,
				"error", ctx.Err(),
			)
			return combined
		}
	}

	return combined
}

func buildQuery(typ string, id string, season int, episode int) string {
	query := normalizeQuery(id)
	if typ == "series" && season > 0 && episode > 0 {
		query = fmt.Sprintf("%s S%02dE%02d", query, season, episode)
	}
	return strings.TrimSpace(query)
}

func normalizeQuery(raw string) string {
	query := strings.ReplaceAll(raw, ".", " ")
	query = strings.ReplaceAll(query, "_", " ")
	query = strings.ReplaceAll(query, "-", " ")
	return strings.TrimSpace(query)
}
