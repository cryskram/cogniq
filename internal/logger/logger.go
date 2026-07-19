package logger

import (
	"io"
	"os"
	"time"

	"github.com/cryskram/cogniq/internal/config"
	"github.com/rs/zerolog"
)

func New(cfg config.LogConfig) zerolog.Logger {
	var output io.Writer = os.Stderr

	if cfg.Output != "" && cfg.Output != "stdout" && cfg.Output != "stderr" {
		f, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

		if err != nil {
			output = f
		}
	}

	if cfg.Output == "stdout" {
		output = os.Stdout
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano

	if cfg.Format == "json" {
		return zerolog.New(output).With().Timestamp().Logger()
	}

	return zerolog.New(zerolog.ConsoleWriter{Out: output, TimeFormat: "15:04:05"}).With().Timestamp().Logger()
}