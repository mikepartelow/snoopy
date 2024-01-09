package logging

import (
	"fmt"
	"os"
	"strings"

	"log/slog"
)

func Init() (*slog.Logger, error) {
	level, err := getLevel()
	if err != nil {
		return nil, fmt.Errorf("couldn't set log level: %w", err)
	}

	var h slog.Handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})

	return slog.New(h), nil
}

func getLevel() (slog.Level, error) {
	level := slog.LevelInfo

	if levelVal := os.Getenv("LOG_LEVEL"); levelVal != "" {
		switch strings.ToLower(levelVal) {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			return 0, fmt.Errorf("unkown log level: " + levelVal)
		}
	}

	return level, nil
}
