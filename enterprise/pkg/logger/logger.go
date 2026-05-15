package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"

	"web-demo/enterprise/config"
)

// New 根据配置创建 zerolog.Logger
func New(cfg config.LogConfig) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	if cfg.Format == "json" {
		return zerolog.New(os.Stdout).
			Level(level).
			With().
			Timestamp().
			Caller().
			Logger()
	}

	return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()
}
