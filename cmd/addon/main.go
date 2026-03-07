package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"scrapper/internal/addon"
)

func main() {
	cfg := addon.DefaultConfig()
	loadConfigFile("config.yml", &cfg)
	applyEnvOverrides(&cfg)

	port := cfg.Port

	logger := slog.New(newColorHandler(os.Stdout, slog.LevelDebug))
	slog.SetDefault(logger)

	app, err := addon.New(cfg, logger)
	if err != nil {
		log.Fatalf("init addon: %v", err)
	}

	mux := http.NewServeMux()
	app.Register(mux)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}
