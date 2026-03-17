package addon

import (
	"context"
	"time"

	"scrapper/internal/indexer"
)

func (a *Addon) searchStreams(ctx context.Context, query indexer.SearchQuery) []indexer.Result {
	// TODO: Get timeout from config
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	type response struct {
		results []indexer.Result
		err     error
		name    string
	}

	ch := make(chan response, len(a.indexers))
	for _, idx := range a.indexers {
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
					"id", query.ID,
					"error", res.err,
				)
				continue
			}
			if len(res.results) == 0 {
				a.logger.Info("indexer search empty",
					"indexer", res.name,
					"id", query.ID,
				)
				continue
			}
			a.logger.Info("indexer search success",
				"indexer", res.name,
				"id", query.ID,
				"results", len(res.results),
			)
			combined = append(combined, res.results...)
		case <-ctx.Done():
			a.logger.Warn("indexer search timeout",
				"id", query.ID,
				"error", ctx.Err(),
			)
			return combined
		}
	}

	return combined
}
