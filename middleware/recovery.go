package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"enmproperty/backend/platform/httpx"
)

// Recovery menangkap panic dan mengembalikan respons 500 standar tanpa
// membocorkan detail internal, sekaligus mencatatnya sebagai error.
func Recovery(logger zerolog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		logger.Error().
			Interface("panic", recovered).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Msg("panic tertangkap pada handler")

		httpx.Fail(c, http.StatusInternalServerError, httpx.CodeInternal, "Terjadi kesalahan pada server")
	})
}

// ErrorHandler memformat error yang dititipkan handler ke konteks menjadi
// respons error standar {code, message, details}.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if err := httpx.Error(c); err != nil && !c.Writer.Written() {
			httpx.Respond(c, err)
		}
	}
}
