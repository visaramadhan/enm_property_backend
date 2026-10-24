package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"enmproperty/backend/models"
)

// AuditEntry adalah aksi admin atau superadmin yang dicatat setelah handler berhasil.
type AuditEntry struct {
	Action   models.AuditAction
	Entity   string
	EntityID string
	Before   string
	After    string
	Note     string
}

// SetAudit menitipkan rencana pencatatan audit yang akan diproses AuditHook.
func SetAudit(c *gin.Context, entry AuditEntry) {
	c.Set(contextAuditKey, entry)
}

// AuditHook mencatat aksi pengelola platform pada audit log setelah handler
// selesai dengan status sukses. Aksi yang gagal tidak dicatat.
func AuditHook(audits AuditWriter) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() >= 400 {
			return
		}

		value, exists := c.Get(contextAuditKey)
		if !exists {
			return
		}
		entry, ok := value.(AuditEntry)
		if !ok {
			return
		}

		user := CurrentUser(c)
		if user == nil || !user.Role.IsInternal() {
			return
		}

		record := &models.AuditLog{
			ID:         uuid.NewString(),
			ActorID:    user.ID,
			ActorName:  user.DisplayName(),
			ActorEmail: user.Email,
			ActorRole:  user.Role,
			Action:     entry.Action,
			Entity:     entry.Entity,
			EntityID:   entry.EntityID,
			Before:     entry.Before,
			After:      entry.After,
			Note:       entry.Note,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			CreatedAt:  time.Now(),
		}
		_ = audits.Create(record)
	}
}
