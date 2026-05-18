package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init configures the global zerolog logger.
//
// env: "dev" → human-readable console output; anything else → JSON.
// level: "debug" | "info" | "warn" | "error". Invalid → falls back to "info".
//
// Output always goes to os.Stderr (Unix convention: logs on stderr, data on stdout).
// Caller info (file:line) is enabled to aid debugging during the bootcamp.
func Init(env, level string) {
	// Parse log level — fallback to Info if invalid
	lvl, err := zerolog.ParseLevel(level)
	if err != nil || lvl == zerolog.NoLevel {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	// Format
	if env == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})
	}
	// else: default zerolog writes JSON to os.Stderr — leave as-is

	// Enable caller (file:line) globally
	log.Logger = log.With().Caller().Timestamp().Logger()
}
