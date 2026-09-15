// Package logger configures the process-wide zerolog logger: structured JSON in
// production, colorized console output in development.
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(env string, level string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	if env == "production" {
		return zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}
	return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}).
		With().Timestamp().Logger()
}
