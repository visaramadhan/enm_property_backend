package database

import (
	"context"
	"errors"

	"enmproperty/backend/models"
)

type memoryUserRepository struct{ store *Store }

func (r *memoryUserRepository) Create(user *models.User) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	for _, existing := range r.store.users {
		if existing.FirebaseUID == user.FirebaseUID && user.FirebaseUID != "" {
			return errors.New("firebase uid sudah terdaftar")
		}
		if existing.Email == user.Email && user.Email != "" {
			return errors.New("email sudah terdaftar")
		}
	}
	r.store.users[user.ID] = copyUser(user)
	return nil
}

func (r *memoryUserRepository) Update(user *models.User) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.users[user.ID]; !exists {
		return ErrNotFound
	}
	r.store.users[user.ID] = copyUser(user)
	return nil
}

func (r *memoryUserRepository) GetByID(id string) (*models.User, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return copyUser(r.store.users[id]), nil
}

func (r *memoryUserRepository) GetByFirebaseUID(uid string) (*models.User, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, user := range r.store.users {
		if user.FirebaseUID == uid {
			return copyUser(user), nil
		}
	}
	return nil, ErrNotFound
}

func (r *memoryUserRepository) GetByEmail(email string) (*models.User, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, user := range r.store.users {
		if equalFold(user.Email, email) {
			return copyUser(user), nil
		}
	}
	return nil, ErrNotFound
}

func (r *memoryUserRepository) List(q ListQuery) ([]*models.User, int, error) {
	items := make([]*models.User, 0)
	r.store.mu.RLock()
	for _, user := range r.store.users {
		items = append(items, copyUser(user))
	}
	r.store.mu.RUnlock()

	filtered := make([]*models.User, 0, len(items))
	for _, user := range items {
		if role := q.Filter("role"); role != "" && string(user.Role) != role {
			continue
		}
		if status := q.Filter("status"); status != "" && string(user.Status) != status {
			continue
		}
		if q.Search != "" && !ContainsFold(user.Name, q.Search) && !ContainsFold(user.Email, q.Search) {
			continue
		}
		filtered = append(filtered, user)
	}

	SortByFunc(filtered, func(a, b *models.User) bool {
		if q.SortBy == "name" {
			return a.Name < b.Name
		}
		return a.CreatedAt.Before(b.CreatedAt)
	}, q.SortBy != "name")

	page, total := Paginate(filtered, q)
	return page, total, nil
}

// DeleteUserCascade menghapus akun beserta seluruh data terkait dari penyimpanan
// in-memory. Query yang bergantung pada partnerID dilewati bila pengguna bukan
// mitra (partnerID kosong).
func (r *memoryUserRepository) DeleteUserCascade(_ context.Context, userID, partnerID string, propertyIDs []string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	propertySet := make(map[string]struct{}, len(propertyIDs))
	for _, id := range propertyIDs {
		propertySet[id] = struct{}{}
	}

	for id, photo := range r.store.photos {
		if _, ok := propertySet[photo.PropertyID]; ok {
			delete(r.store.photos, id)
		}
	}

	views := r.store.views[:0]
	for _, view := range r.store.views {
		if _, ok := propertySet[view.PropertyID]; !ok {
			views = append(views, view)
		}
	}
	r.store.views = views

	contacts := r.store.contacts[:0]
	for _, contact := range r.store.contacts {
		if _, ok := propertySet[contact.PropertyID]; !ok {
			contacts = append(contacts, contact)
		}
	}
	r.store.contacts = contacts

	for id, favorite := range r.store.favorites {
		_, byProperty := propertySet[favorite.PropertyID]
		if favorite.UserID == userID || byProperty {
			delete(r.store.favorites, id)
		}
	}

	for id, property := range r.store.properties {
		if property.PartnerID == partnerID && partnerID != "" {
			delete(r.store.properties, id)
		}
	}
	for id, subscription := range r.store.subscriptions {
		if subscription.PartnerID == partnerID && partnerID != "" {
			delete(r.store.subscriptions, id)
		}
	}
	for id, order := range r.store.orders {
		if order.PartnerID == partnerID && partnerID != "" {
			delete(r.store.orders, id)
		}
	}
	for id, acceptance := range r.store.acceptances {
		if acceptance.UserID == userID || (partnerID != "" && acceptance.PartnerID == partnerID) {
			delete(r.store.acceptances, id)
		}
	}
	for id, partner := range r.store.partners {
		if partner.UserID == userID {
			delete(r.store.partners, id)
		}
	}
	delete(r.store.users, userID)
	return nil
}

type memoryPartnerRepository struct{ store *Store }

func (r *memoryPartnerRepository) Create(partner *models.Partner) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.partners[partner.ID] = copyPartner(partner)
	return nil
}

func (r *memoryPartnerRepository) Update(partner *models.Partner) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.partners[partner.ID]; !exists {
		return ErrNotFound
	}
	r.store.partners[partner.ID] = copyPartner(partner)
	return nil
}

func (r *memoryPartnerRepository) GetByID(id string) (*models.Partner, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	partner, exists := r.store.partners[id]
	if !exists {
		return nil, ErrNotFound
	}
	return copyPartner(partner), nil
}

func (r *memoryPartnerRepository) GetByUserID(userID string) (*models.Partner, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, partner := range r.store.partners {
		if partner.UserID == userID {
			return copyPartner(partner), nil
		}
	}
	return nil, ErrNotFound
}

func (r *memoryPartnerRepository) List(q ListQuery) ([]*models.Partner, int, error) {
	items := make([]*models.Partner, 0)
	r.store.mu.RLock()
	for _, partner := range r.store.partners {
		items = append(items, copyPartner(partner))
	}
	r.store.mu.RUnlock()

	filtered := make([]*models.Partner, 0, len(items))
	for _, partner := range items {
		if status := q.Filter("status"); status != "" && string(partner.Status) != status {
			continue
		}
		if city := q.Filter("city"); city != "" && !equalFold(partner.City, city) {
			continue
		}
		if q.Search != "" && !ContainsFold(partner.BusinessName, q.Search) &&
			!ContainsFold(partner.PICName, q.Search) && !ContainsFold(partner.Email, q.Search) {
			continue
		}
		filtered = append(filtered, partner)
	}

	SortByFunc(filtered, func(a, b *models.Partner) bool {
		return a.CreatedAt.Before(b.CreatedAt)
	}, true)

	page, total := Paginate(filtered, q)
	return page, total, nil
}

type memoryPlanRepository struct{ store *Store }

func (r *memoryPlanRepository) Create(plan *models.Plan) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	for _, existing := range r.store.plans {
		if existing.Slug == plan.Slug {
			return errors.New("slug paket sudah dipakai")
		}
	}
	r.store.plans[plan.ID] = copyPlan(plan)
	return nil
}

func (r *memoryPlanRepository) Update(plan *models.Plan) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.plans[plan.ID]; !exists {
		return ErrNotFound
	}
	r.store.plans[plan.ID] = copyPlan(plan)
	return nil
}

func (r *memoryPlanRepository) GetByID(id string) (*models.Plan, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	plan, exists := r.store.plans[id]
	if !exists {
		return nil, ErrNotFound
	}
	return copyPlan(plan), nil
}

func (r *memoryPlanRepository) GetBySlug(slug string) (*models.Plan, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, plan := range r.store.plans {
		if plan.Slug == slug {
			return copyPlan(plan), nil
		}
	}
	return nil, ErrNotFound
}

func (r *memoryPlanRepository) List(q ListQuery) ([]*models.Plan, int, error) {
	items := make([]*models.Plan, 0)
	r.store.mu.RLock()
	for _, plan := range r.store.plans {
		items = append(items, copyPlan(plan))
	}
	r.store.mu.RUnlock()

	filtered := make([]*models.Plan, 0, len(items))
	for _, plan := range items {
		if q.Filter("is_active") == "true" && !plan.IsActive {
			continue
		}
		if q.Filter("is_active") == "false" && plan.IsActive {
			continue
		}
		if q.Search != "" && !ContainsFold(plan.Name, q.Search) {
			continue
		}
		filtered = append(filtered, plan)
	}

	SortByFunc(filtered, func(a, b *models.Plan) bool {
		return a.SortOrder < b.SortOrder
	}, false)

	page, total := Paginate(filtered, q)
	return page, total, nil
}
