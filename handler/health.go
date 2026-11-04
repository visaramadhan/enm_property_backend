package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// HealthHandler menyediakan pemeriksaan kesehatan dan kesiapan layanan.
type HealthHandler struct {
	services *services.Services
	started  time.Time
}

// NewHealthHandler membuat handler kesehatan dengan penanda waktu mulai.
func NewHealthHandler(svc *services.Services) *HealthHandler {
	return &HealthHandler{services: svc, started: time.Now()}
}

// Health mengembalikan status dasar layanan.
func (h *HealthHandler) Health(c *gin.Context) {
	httpx.OK(c, gin.H{
		"status":  "ok",
		"service": "enmproperty-api",
		"time":    time.Now(),
	})
}

// Ready menandakan layanan siap menerima permintaan beserta lama berjalan.
func (h *HealthHandler) Ready(c *gin.Context) {
	httpx.OK(c, gin.H{
		"status":         "ready",
		"uptime_seconds": int(time.Since(h.started).Seconds()),
	})
}
