package database

import (
	"strings"
	"time"

	"enmproperty/backend/models"
)

// equalFold membandingkan dua teks tanpa memandang huruf besar-kecil.
func equalFold(a, b string) bool { return strings.EqualFold(a, b) }

// withinRange memeriksa apakah waktu berada dalam filter from/to (RFC3339).
func withinRange(t time.Time, q ListQuery) bool {
	if from := q.Filter("from"); from != "" {
		if parsed, err := time.Parse(time.RFC3339, from); err == nil && t.Before(parsed) {
			return false
		}
	}
	if to := q.Filter("to"); to != "" {
		if parsed, err := time.Parse(time.RFC3339, to); err == nil && t.After(parsed) {
			return false
		}
	}
	return true
}

type memorySubscriptionRepository struct{ store *Store }

func (r *memorySubscriptionRepository) Create(subscription *models.Subscription) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.subscriptions[subscription.ID] = copySubscription(subscription)
	return nil
}

func (r *memorySubscriptionRepository) Update(subscription *models.Subscription) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.subscriptions[subscription.ID]; !exists {
		return ErrNotFound
	}
	r.store.subscriptions[subscription.ID] = copySubscription(subscription)
	return nil
}

func (r *memorySubscriptionRepository) GetByID(id string) (*models.Subscription, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	subscription, exists := r.store.subscriptions[id]
	if !exists {
		return nil, ErrNotFound
	}
	return copySubscription(subscription), nil
}

// GetActiveByPartner mengembalikan langganan dengan EndsAt terjauh milik mitra.
func (r *memorySubscriptionRepository) GetActiveByPartner(partnerID string) (*models.Subscription, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	var newest *models.Subscription
	for _, subscription := range r.store.subscriptions {
		if subscription.PartnerID != partnerID {
			continue
		}
		if newest == nil || subscription.EndsAt.After(newest.EndsAt) {
			newest = subscription
		}
	}
	if newest == nil {
		return nil, ErrNotFound
	}
	return copySubscription(newest), nil
}

func (r *memorySubscriptionRepository) GetLatestByPartner(partnerID string) (*models.Subscription, error) {
	return r.GetActiveByPartner(partnerID)
}

func (r *memorySubscriptionRepository) List(q ListQuery) ([]*models.Subscription, int, error) {
	items := make([]*models.Subscription, 0)
	r.store.mu.RLock()
	for _, subscription := range r.store.subscriptions {
		items = append(items, copySubscription(subscription))
	}
	r.store.mu.RUnlock()

	filtered := make([]*models.Subscription, 0, len(items))
	for _, subscription := range items {
		if partnerID := q.Filter("partner_id"); partnerID != "" && subscription.PartnerID != partnerID {
			continue
		}
		if planID := q.Filter("plan_id"); planID != "" && subscription.PlanID != planID {
			continue
		}
		if status := q.Filter("status"); status != "" && string(subscription.Status) != status {
			continue
		}
		if !withinRange(subscription.CreatedAt, q) {
			continue
		}
		filtered = append(filtered, subscription)
	}

	SortByFunc(filtered, func(a, b *models.Subscription) bool {
		return a.CreatedAt.Before(b.CreatedAt)
	}, true)

	page, total := Paginate(filtered, q)
	return page, total, nil
}

type memoryOrderRepository struct{ store *Store }

func (r *memoryOrderRepository) Create(order *models.Order) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	for _, existing := range r.store.orders {
		if existing.Code == order.Code {
			return errDuplicateCode
		}
	}
	r.store.orders[order.ID] = copyOrder(order)
	return nil
}

func (r *memoryOrderRepository) Update(order *models.Order) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.orders[order.ID]; !exists {
		return ErrNotFound
	}
	r.store.orders[order.ID] = copyOrder(order)
	return nil
}

func (r *memoryOrderRepository) GetByID(id string) (*models.Order, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	order, exists := r.store.orders[id]
	if !exists {
		return nil, ErrNotFound
	}
	return copyOrder(order), nil
}

func (r *memoryOrderRepository) GetByCode(code string) (*models.Order, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, order := range r.store.orders {
		if equalFold(order.Code, code) {
			return copyOrder(order), nil
		}
	}
	return nil, ErrNotFound
}

func (r *memoryOrderRepository) CountByCodePrefix(prefix string) int {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	count := 0
	for _, order := range r.store.orders {
		if strings.HasPrefix(order.Code, prefix) {
			count++
		}
	}
	return count
}

func (r *memoryOrderRepository) List(q ListQuery) ([]*models.Order, int, error) {
	items := make([]*models.Order, 0)
	r.store.mu.RLock()
	for _, order := range r.store.orders {
		items = append(items, copyOrder(order))
	}
	r.store.mu.RUnlock()

	filtered := make([]*models.Order, 0, len(items))
	for _, order := range items {
		if status := q.Filter("status"); status != "" && string(order.Status) != status {
			continue
		}
		if partnerID := q.Filter("partner_id"); partnerID != "" && order.PartnerID != partnerID {
			continue
		}
		if userID := q.Filter("user_id"); userID != "" && order.UserID != userID {
			continue
		}
		if planID := q.Filter("plan_id"); planID != "" && order.PlanID != planID {
			continue
		}
		if provider := q.Filter("provider"); provider != "" && string(order.Provider) != provider {
			continue
		}
		if kind := q.Filter("kind"); kind != "" && string(order.Kind) != kind {
			continue
		}
		if !withinRange(order.CreatedAt, q) {
			continue
		}
		if q.Search != "" && !ContainsFold(order.Code, q.Search) &&
			!ContainsFold(order.PlanName, q.Search) && !ContainsFold(order.PartnerName, q.Search) {
			continue
		}
		filtered = append(filtered, order)
	}

	SortByFunc(filtered, func(a, b *models.Order) bool {
		return a.CreatedAt.Before(b.CreatedAt)
	}, true)

	page, total := Paginate(filtered, q)
	return page, total, nil
}

// errDuplicateCode dikembalikan bila kode order sudah terpakai.
var errDuplicateCode = duplicateCodeError{}

type duplicateCodeError struct{}

func (duplicateCodeError) Error() string { return "kode order sudah terpakai" }

// IsDuplicateCode memeriksa error kode order duplikat untuk percobaan ulang penomoran.
func IsDuplicateCode(err error) bool { return err == errDuplicateCode }
