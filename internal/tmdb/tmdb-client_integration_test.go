package tmdb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"scrapper/internal/config"
)

// newLiveClient loads the repo-root .env, skips the test when no TMDB_API_KEY is
// available, and returns a client pointed at the real TMDB API.
func newLiveClient(t *testing.T) *Client {
	t.Helper()

	// Ignore the error: the key may already be exported in the environment.
	_ = godotenv.Load("../../.env")

	if os.Getenv("TMDB_API_KEY") == "" {
		t.Skip("TMDB_API_KEY not set; skipping live TMDB integration test")
	}

	cfg := config.TMDBConfig{
		BaseURL: "https://api.themoviedb.org/3",
		Timeout: 30,
	}

	client, err := NewClient(cfg, nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return client
}

func TestTVTitleByIMDB_Integration(t *testing.T) {
	client := newLiveClient(t)

	t.Run("known TV series", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		results, err := client.TVTitleByIMDB(ctx, "tt0903747") // Breaking Bad
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected at least one tv result, got none")
		}
		if results[0].Name != "Breaking Bad" {
			t.Fatalf("expected first result %q, got %q", "Breaking Bad", results[0].Name)
		}
	})

	t.Run("whitespace-padded id is trimmed", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		results, err := client.TVTitleByIMDB(ctx, "  tt0944947  ") // Game of Thrones
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected at least one tv result, got none")
		}
		if results[0].Name != "Game of Thrones" {
			t.Fatalf("expected first result %q, got %q", "Game of Thrones", results[0].Name)
		}
	})

	t.Run("movie id yields no tv results", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		results, err := client.TVTitleByIMDB(ctx, "tt0111161") // The Shawshank Redemption (film)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("expected no tv results for a movie id, got %v", results)
		}
	})

	t.Run("nonexistent id yields no results", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		results, err := client.TVTitleByIMDB(ctx, "tt9999999999")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("expected no results for a nonexistent id, got %v", results)
		}
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel before the request is issued

		results, err := client.TVTitleByIMDB(ctx, "tt0903747")
		if err == nil {
			t.Fatalf("expected an error for a cancelled context, got results %v", results)
		}
	})
}
