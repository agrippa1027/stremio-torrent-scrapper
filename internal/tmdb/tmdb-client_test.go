package tmdb

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/charmbracelet/log"

	"scrapper/internal/config"
)

func testLogger() *log.Logger {
	return log.New(io.Discard)
}

func testConfig() config.TMDBConfig {
	return config.TMDBConfig{
		BaseURL: "https://api.themoviedb.org/3",
		Timeout: 30,
	}
}

func TestNewClient_MissingAPIKey(t *testing.T) {
	t.Setenv("TMDB_API_KEY", "")

	client, err := New(testConfig(), nil, testLogger())
	if !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("expected ErrMissingAPIKey, got %v", err)
	}
	if client != nil {
		t.Fatalf("expected nil client, got %v", client)
	}
}

func TestNewClient_Success(t *testing.T) {
	t.Setenv("TMDB_API_KEY", "dummy")

	client, err := New(testConfig(), nil, testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.apiKey != "dummy" {
		t.Fatalf("expected apiKey %q, got %q", "dummy", client.apiKey)
	}
}

func TestTVTitleByIMDB_EmptyID(t *testing.T) {
	t.Setenv("TMDB_API_KEY", "dummy")

	client, err := New(testConfig(), nil, testLogger())
	if err != nil {
		t.Fatalf("unexpected error building client: %v", err)
	}

	// The empty-id guard fires before any HTTP call, so this stays offline.
	results, err := client.Title(context.Background(), "   ")
	if err == nil {
		t.Fatal("expected an error for empty imdb id, got nil")
	}
	if results != nil {
		t.Fatalf("expected nil results, got %v", results)
	}
}
