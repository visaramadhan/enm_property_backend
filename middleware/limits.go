package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"

	"enmproperty/backend/platform/httpx"
)

// BodyLimit membatasi ukuran body permintaan agar tidak melebihi batas aman.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

// Timeout membatalkan permintaan yang melebihi durasi tertentu dengan respons 408.
func Timeout(duration time.Duration) gin.HandlerFunc {
	if duration <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return timeout.New(
		timeout.WithTimeout(duration),
		timeout.WithResponse(func(c *gin.Context) {
			httpx.Fail(c, http.StatusRequestTimeout, httpx.CodeBadRequest, "Permintaan melebihi batas waktu")
		}),
	)
}
