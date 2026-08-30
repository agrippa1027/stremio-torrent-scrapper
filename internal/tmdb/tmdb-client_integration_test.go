package tmdb

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
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

	opt := Options{
		BaseURL:        "https://api.themoviedb.org/3",
		APIKey:         os.Getenv("TMDB_API_KEY"),
		TimeoutSeconds: 30,
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSHandshakeTimeout = 20 * time.Second

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	client, err := New(opt, httpClient)
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

		results, err := client.TitleByIMDB(ctx, "tt0903747") // Breaking Bad
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if results != "Breaking Bad" {
			t.Fatalf("expected first result %q, got %q", "Breaking Bad", results)
		}
	})

	t.Run("whitespace-padded id is trimmed", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		results, err := client.TitleByIMDB(ctx, "  tt0944947  ") // Game of Thrones
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if results != "Game of Thrones" {
			t.Fatalf("expected first result %q, got %q", "Game of Thrones", results)
		}
	})

	t.Run("movie id yields no tv results", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		results, err := client.TitleByIMDB(ctx, "tt0111161") // The Shawshank Redemption (film)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if results != "The Shawshank Redemption" {
			t.Fatalf("expected first result %q, got %q", "The Shawshank Redemption", results)
		}
	})

	t.Run("nonexistent id yields no results", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		results, err := client.TitleByIMDB(ctx, "tt9999999999")
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

		results, err := client.TitleByIMDB(ctx, "tt0903747")
		if err == nil {
			t.Fatalf("expected an error for a cancelled context, got results %v", results)
		}
	})
}
