package middleware

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	"enmproperty/backend/platform/logger"
)

// RequestID menambahkan identitas unik pada setiap permintaan untuk penelusuran log.
func RequestID() gin.HandlerFunc {
	return requestid.New(
		requestid.WithHandler(func(c *gin.Context, requestID string) {
			c.Set(logger.RequestIDKey, requestID)
			c.Header("X-Request-Id", requestID)
		}),
	)
}
