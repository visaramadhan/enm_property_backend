// Package logger menyiapkan zerolog sebagai logger terstruktur aplikasi.
package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"enmproperty/backend/config"
)

// New membuat logger zerolog sesuai konfigurasi LOG_LEVEL dan LOG_FORMAT.
func New(cfg config.Log) zerolog.Logger {
	level := parseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var writer io.Writer = os.Stdout
	if strings.EqualFold(cfg.Format, "console") {
		writer = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	return zerolog.New(writer).
		With().
		Timestamp().
		Str("service", "enmproperty-api").
		Logger().
		Level(level)
}

// RequestIDKey adalah kunci konteks Gin tempat request id disimpan.
const RequestIDKey = "request_id"

// From mengembalikan logger yang sudah dilengkapi request id pada konteks.
func From(c *gin.Context, base zerolog.Logger) zerolog.Logger {
	if value, exists := c.Get(RequestIDKey); exists {
		if id, ok := value.(string); ok && id != "" {
			return base.With().Str("request_id", id).Logger()
		}
	}
	return base
}

// parseLevel mengubah nama level menjadi zerolog.Level dengan fallback info.
func parseLevel(raw string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return zerolog.DebugLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
