package main

import (
	"net/http"
	"os"
	"time"

	"scrapper/internal/addon"

	"github.com/charmbracelet/log"
)


func main() {
	cfg := addon.DefaultConfig()
	loadConfigFile("config.yml", &cfg)

	port := cfg.Port

	logger := log.New(os.Stdout)
	logger.SetLevel(log.DebugLevel)


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

	logger.Info("listening on port", "port", port)
	logger.Warn(srv.ListenAndServe())
}
