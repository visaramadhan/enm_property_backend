package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	memorystore "github.com/ulule/limiter/v3/drivers/store/memory"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"

	"enmproperty/backend/config"
	"enmproperty/backend/platform/httpx"
)

// RateLimit membatasi jumlah permintaan per IP dalam satu jendela waktu.
// Nilai requests <= 0 atau rate limit yang dimatikan mengembalikan no-op.
func RateLimit(cfg config.RateLimit, requests int) gin.HandlerFunc {
	if !cfg.Enabled || requests <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	rate := limiter.Rate{
		Period: cfg.Window,
		Limit:  int64(requests),
	}
	instance := limiter.New(memorystore.NewStore(), rate)

	return ginlimiter.NewMiddleware(instance,
		ginlimiter.WithLimitReachedHandler(func(c *gin.Context) {
			httpx.Fail(c, http.StatusTooManyRequests, httpx.CodeTooManyRequest, "Terlalu banyak permintaan, coba lagi nanti")
		}),
		ginlimiter.WithErrorHandler(func(c *gin.Context, err error) {
			httpx.Respond(c, httpx.Internal(err))
		}),
	)
}
