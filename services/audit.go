package services

import (
	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
)

// AuditService mencatat dan membaca audit log aksi pengelola platform (BE-15).
type AuditService struct {
	deps Deps
}

// Log menyimpan satu catatan audit yang sudah terbentuk.
func (s *AuditService) Log(entry *models.AuditLog) error {
	if entry == nil {
		return nil
	}
	if entry.ID == "" {
		entry.ID = newID()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = timeNow()
	}
	return s.deps.Repos.Audits.Create(entry)
}

// Record membentuk catatan audit dari aksi pengelola platform. Kesalahan
// penyimpanan diabaikan agar tidak menggagalkan permintaan utama.
func (s *AuditService) Record(actor *models.User, action models.AuditAction, entity, entityID string, before, after any, note string, meta RequestMeta) {
	entry := &models.AuditLog{
		ID:        newID(),
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Before:    toJSON(before),
		After:     toJSON(after),
		Note:      note,
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
		CreatedAt: timeNow(),
	}
	if actor != nil {
		entry.ActorID = actor.ID
		entry.ActorName = actor.DisplayName()
		entry.ActorEmail = actor.Email
		entry.ActorRole = actor.Role
	}
	_ = s.Log(entry)
}

// List mengembalikan daftar audit log terfilter untuk halaman admin.
func (s *AuditService) List(q dto.AuditFilterQuery) (dto.ListResult[dto.AuditView], error) {
	entries, total, err := s.deps.Repos.Audits.List(auditListQuery(q))
	if err != nil {
		return dto.ListResult[dto.AuditView]{}, httpx.Internal(err)
	}
	return dto.NewListResult(dto.NewAuditViews(entries), total), nil
}

// Export mengembalikan seluruh audit log sesuai filter untuk diekspor.
func (s *AuditService) Export(q dto.AuditFilterQuery) ([]dto.AuditView, error) {
	query := auditListQuery(q)
	query.Page = 0
	query.PerPage = 0
	entries, _, err := s.deps.Repos.Audits.List(query)
	if err != nil {
		return nil, httpx.Internal(err)
	}
	return dto.NewAuditViews(entries), nil
}

// auditListQuery memetakan filter permintaan menjadi ListQuery repository.
func auditListQuery(q dto.AuditFilterQuery) database.ListQuery {
	filters := newFilters()
	putFilter(filters, "entity", q.Entity)
	putFilter(filters, "entity_id", q.EntityID)
	putFilter(filters, "actor_id", q.ActorID)
	putFilter(filters, "actor_role", q.ActorRole)
	putFilter(filters, "action", q.Action)
	putFilter(filters, "from", q.From)
	putFilter(filters, "to", q.To)
	return database.ListQuery{
		Search:  q.Search,
		Page:    q.Page,
		PerPage: q.PerPage,
		Filters: filters,
	}
}
