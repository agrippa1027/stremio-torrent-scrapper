package infrastructure

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
)

func logger(level string) *log.Logger {
	logLevel, err := log.ParseLevel(level)

	if err != nil {
		log.Fatal("parse log level: %v", err)
	}

	logger := log.New(os.Stdout)
	logger.SetLevel(logLevel)
	return logger
}

func WithLogger(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, "loggerKey", logger)
}

func FromContext(ctx context.Context) *log.Logger {
	if logger, ok := ctx.Value("loggerKey").(*log.Logger); ok {
		return logger
	}
	return logger(os.Getenv("LOG_LEVEL"))
}
