package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"enmproperty/backend/platform/logger"
)

// AccessLogger mencatat method, path, status, dan durasi setiap permintaan
// tanpa menyertakan body maupun header sensitif.
func AccessLogger(base zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		event := base.Info()
		if c.Writer.Status() >= 500 {
			event = base.Error()
		} else if c.Writer.Status() >= 400 {
			event = base.Warn()
		}

		event.
			Str("request_id", requestID(c)).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("duration", time.Since(start)).
			Str("ip", c.ClientIP()).
			Msg("akses")
	}
}

// requestID mengambil request id dari konteks untuk kebutuhan log.
func requestID(c *gin.Context) string {
	if value, exists := c.Get(logger.RequestIDKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}
