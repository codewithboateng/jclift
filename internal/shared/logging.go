package shared

import (
	"log/slog"
	"os"
	"strings"
)

func InitLogger(format, level string) *slog.Logger {
	var h slog.Handler
	lvl := slog.LevelInfo
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}
	if strings.ToLower(format) == "text" {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	} else {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	}
	logger := slog.New(h)
	slog.SetDefault(logger)
	return logger
}
