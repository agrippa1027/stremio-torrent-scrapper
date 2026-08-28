package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/charmbracelet/log"

	"scrapper/internal/config"
	"scrapper/internal/stremio"
	"scrapper/logging"
)

func main() {
	loadEnv()

	cfg, err := config.LoadConfigFile("config.yml")
	if err != nil {
		log.Fatal("load config: %v", err)
		return
	}

	// TODO: check on this
	// check what? dumb fucking comments
	ctx := context.Background()
	logger := logging.FromContext(ctx)
	port := os.Getenv("PORT")

	app, err := stremio.New(cfg, logger)
	if err != nil {
		logger.Fatal("init addon: %v", err)
		return
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
	logger.Fatal("listen and serve", "err", srv.ListenAndServe())
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Warn("load env from file: %v", err)
	}
}
