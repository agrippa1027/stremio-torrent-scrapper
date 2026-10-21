// NOTE:
// This is the logger package
// currently uses charmbracelet/log
// but should re-think the decision to use the charmbracelet/log
// it would be better/simpler to use a standard logging library
package logging

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
)

func Logger(level string) *log.Logger {
	logLevel, err := log.ParseLevel(level)

	if err != nil {
		log.Fatalf("parse log level: %v", err)
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
