package database

import (
	"enmproperty/backend/models"
)

type memoryAuditRepository struct{ store *Store }

func (r *memoryAuditRepository) Create(entry *models.AuditLog) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.audits = append(r.store.audits, copyAudit(entry))
	return nil
}

func (r *memoryAuditRepository) List(q ListQuery) ([]*models.AuditLog, int, error) {
	items := r.All()

	filtered := make([]*models.AuditLog, 0, len(items))
	for _, entry := range items {
		if entity := q.Filter("entity"); entity != "" && !equalFold(entry.Entity, entity) {
			continue
		}
		if entityID := q.Filter("entity_id"); entityID != "" && entry.EntityID != entityID {
			continue
		}
		if actorID := q.Filter("actor_id"); actorID != "" && entry.ActorID != actorID {
			continue
		}
		if actorRole := q.Filter("actor_role"); actorRole != "" && string(entry.ActorRole) != actorRole {
			continue
		}
		if action := q.Filter("action"); action != "" && string(entry.Action) != action {
			continue
		}
		if !withinRange(entry.CreatedAt, q) {
			continue
		}
		if q.Search != "" && !ContainsFold(entry.ActorName, q.Search) &&
			!ContainsFold(entry.ActorEmail, q.Search) && !ContainsFold(entry.Note, q.Search) {
			continue
		}
		filtered = append(filtered, entry)
	}

	SortByFunc(filtered, func(a, b *models.AuditLog) bool {
		return a.CreatedAt.Before(b.CreatedAt)
	}, true)

	page, total := Paginate(filtered, q)
	return page, total, nil
}

func (r *memoryAuditRepository) All() []*models.AuditLog {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	audits := make([]*models.AuditLog, 0, len(r.store.audits))
	for _, entry := range r.store.audits {
		audits = append(audits, copyAudit(entry))
	}
	return audits
}
