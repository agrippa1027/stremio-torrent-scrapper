package main

import (
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"scrapper/internal/addon"
	"scrapper/internal/addon/config"

	"github.com/charmbracelet/log"
)


func main() {
	loadEnv()

	cfg ,err := config.LoadConfigFile("config.yml")
	if err != nil {
		log.Fatal("load config: %v", err)
		return
	}

	logger := logger(os.Getenv("LOG_LEVEL"))

	port := os.Getenv("PORT")

	app, err := addon.New(cfg, logger)
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


func logger(level string) *log.Logger {
	logLevel, err := log.ParseLevel(level)

	if err != nil {
		log.Fatal("parse log level: %v", err)
	}

	logger := log.New(os.Stdout)
	logger.SetLevel(logLevel)
	return logger
}


func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Warn("load env from file: %v", err)
	}
}
