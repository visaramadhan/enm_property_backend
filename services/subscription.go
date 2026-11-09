package services

import (
	"time"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
)

// SubscriptionService mengelola masa aktif paket mitra, perpanjangan, upgrade,
// serta penandaan kedaluwarsa (BE-07/BE-16).
type SubscriptionService struct {
	deps     Deps
	settings *SettingService
}

// GetCurrent mengembalikan langganan berjalan mitra beserta sisa hari, flag
// alert banner, dan pemakaian kuota listing.
func (s *SubscriptionService) GetCurrent(partner *models.Partner) (*dto.SubscriptionView, error) {
	if partner == nil {
		return nil, nil
	}

	subscription, err := s.deps.Repos.Subscriptions.GetActiveByPartner(partner.ID)
	if err != nil {
		if database.IsNotFound(err) {
			return nil, nil
		}
		return nil, httpx.Internal(err)
	}
	view := s.view(subscription)
	return &view, nil
}

// GetByPartner mengembalikan langganan terbaru mitra.
func (s *SubscriptionService) GetByPartner(partnerID string) (*models.Subscription, error) {
	subscription, err := s.deps.Repos.Subscriptions.GetActiveByPartner(partnerID)
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Langganan tidak ditemukan")
		}
		return nil, httpx.Internal(err)
	}
	return subscription, nil
}

// Active mengembalikan langganan terbaru mitra tanpa galat bila belum ada.
func (s *SubscriptionService) Active(partnerID string) *models.Subscription {
	subscription, err := s.deps.Repos.Subscriptions.GetActiveByPartner(partnerID)
	if err != nil {
		return nil
	}
	return subscription
}

// CalcEndDate menghitung akhir masa aktif untuk order baru, perpanjangan, atau upgrade.
func (s *SubscriptionService) CalcEndDate(order *models.Order, now time.Time) time.Time {
	duration := order.PlanDurationDays
	if duration <= 0 {
		duration = 30
	}
	if order.Kind == models.OrderKindRenewal {
		if existing, err := s.deps.Repos.Subscriptions.GetActiveByPartner(order.PartnerID); err == nil && existing.IsActiveAt(now) {
			return existing.EndsAt.AddDate(0, 0, duration)
		}
	}
	return now.AddDate(0, 0, duration)
}

// ActivateFromOrder mengaktifkan atau memperpanjang langganan dari order yang lunas.
func (s *SubscriptionService) ActivateFromOrder(order *models.Order) (*models.Subscription, error) {
	if order == nil {
		return nil, httpx.BadRequest("Order tidak valid")
	}

	now := timeNow()
	existing, err := s.deps.Repos.Subscriptions.GetActiveByPartner(order.PartnerID)
	if err != nil && !database.IsNotFound(err) {
		return nil, httpx.Internal(err)
	}

	duration := order.PlanDurationDays
	if duration <= 0 {
		duration = 30
	}

	// Perpanjangan memperpanjang masa aktif yang masih berjalan.
	if order.Kind == models.OrderKindRenewal && existing != nil && existing.IsActiveAt(now) {
		existing.EndsAt = existing.EndsAt.AddDate(0, 0, duration)
		existing.PlanID = order.PlanID
		existing.PlanName = order.PlanName
		existing.OrderID = order.ID
		existing.Status = models.SubscriptionStatusActive
		existing.UpdatedAt = now
		if err := s.deps.Repos.Subscriptions.Update(existing); err != nil {
			return nil, httpx.Internal(err)
		}
		return existing, nil
	}

	// Upgrade menutup langganan lama lalu membuka periode baru.
	if order.Kind == models.OrderKindUpgrade && existing != nil {
		existing.Status = models.SubscriptionStatusCancelled
		existing.UpdatedAt = now
		if err := s.deps.Repos.Subscriptions.Update(existing); err != nil {
			return nil, httpx.Internal(err)
		}
	}

	subscription := &models.Subscription{
		ID:        newID(),
		PartnerID: order.PartnerID,
		PlanID:    order.PlanID,
		PlanName:  order.PlanName,
		OrderID:   order.ID,
		StartsAt:  now,
		EndsAt:    now.AddDate(0, 0, duration),
		Status:    models.SubscriptionStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.deps.Repos.Subscriptions.Create(subscription); err != nil {
		return nil, httpx.Internal(err)
	}
	return subscription, nil
}

// ExpireDue menandai langganan yang sudah lewat masa aktif menjadi kedaluwarsa,
// menonaktifkan mitra, dan menonaktifkan properti aktifnya. Idempoten.
func (s *SubscriptionService) ExpireDue(now time.Time) (int, error) {
	subscriptions, _, err := s.deps.Repos.Subscriptions.List(database.ListQuery{
		PerPage: 0,
		Filters: map[string]string{"status": string(models.SubscriptionStatusActive)},
	})
	if err != nil {
		return 0, httpx.Internal(err)
	}

	expired := 0
	for _, subscription := range subscriptions {
		if now.Before(subscription.EndsAt) {
			continue
		}

		subscription.Status = models.SubscriptionStatusExpired
		subscription.UpdatedAt = now
		if err := s.deps.Repos.Subscriptions.Update(subscription); err != nil {
			return expired, httpx.Internal(err)
		}

		s.expirePartner(subscription.PartnerID, now)
		s.deps.Repos.Properties.UpdateStatusByPartner(subscription.PartnerID, models.PropertyStatusActive, models.PropertyStatusInactive)
		expired++
	}
	return expired, nil
}

// RestoreProperties mengaktifkan kembali properti mitra yang sempat nonaktif
// karena masa aktif berakhir.
func (s *SubscriptionService) RestoreProperties(partnerID string) int {
	restored := s.deps.Repos.Properties.UpdateStatusByPartner(partnerID, models.PropertyStatusInactive, models.PropertyStatusActive)
	restored += s.deps.Repos.Properties.UpdateStatusByPartner(partnerID, models.PropertyStatusExpired, models.PropertyStatusActive)
	return restored
}

// ListForAdmin mengembalikan daftar langganan untuk kebutuhan laporan.
func (s *SubscriptionService) ListForAdmin(q dto.PaginationQuery, filters map[string]string) (dto.ListResult[dto.SubscriptionView], error) {
	subscriptions, total, err := s.deps.Repos.Subscriptions.List(q.ToListQuery(filters))
	if err != nil {
		return dto.ListResult[dto.SubscriptionView]{}, httpx.Internal(err)
	}

	views := make([]dto.SubscriptionView, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		views = append(views, s.view(subscription))
	}
	return dto.NewListResult(views, total), nil
}

// view melengkapi tampilan langganan dengan sisa hari, flag alert, dan kuota.
func (s *SubscriptionService) view(subscription *models.Subscription) dto.SubscriptionView {
	now := timeNow()
	view := dto.NewSubscriptionView(subscription)
	view.DaysRemaining = subscription.DaysRemaining(now)
	view.IsActive = subscription.IsActiveAt(now)

	threshold := s.deps.Config.App.SubscriptionReminderDays
	if s.settings != nil {
		threshold = s.settings.AlertThresholdDays()
	}
	view.AlertThresholdDays = threshold
	view.ShowAlert = view.IsActive && view.DaysRemaining <= threshold

	quota := -1
	if plan, err := s.deps.Repos.Plans.GetByID(subscription.PlanID); err == nil {
		quota = plan.ListingQuota
		view.BoostQuota = plan.BoostQuota
	}
	used := s.usage(subscription.PartnerID)
	view.ListingQuota = quota
	view.ListingsUsed = used
	view.ListingRemaining = remainingQuota(quota, used)
	view.BoostUsed = s.boostUsage(subscription.PartnerID)

	if order, err := s.deps.Repos.Orders.GetByID(subscription.OrderID); err == nil {
		view.LastOrderCode = order.Code
	}
	return view
}

// usage menghitung jumlah properti mitra yang belum dihapus.
func (s *SubscriptionService) usage(partnerID string) int {
	return s.deps.Repos.Properties.CountByPartner(partnerID)
}

// boostUsage menghitung listing mitra yang sedang di-boost dan belum
// kedaluwarsa, dipakai untuk menampilkan sisa kuota boost paket di sisi mitra.
func (s *SubscriptionService) boostUsage(partnerID string) int {
	properties, _, err := s.deps.Repos.Properties.List(database.ListQuery{
		Filters: map[string]string{"partner_id": partnerID, "boost_active": "true"},
	})
	if err != nil {
		return 0
	}
	now := timeNow()
	count := 0
	for _, property := range properties {
		if property.BoostUntil == nil || property.BoostUntil.After(now) {
			count++
		}
	}
	return count
}

// expirePartner menandai mitra menjadi kedaluwarsa bila sedang aktif.
func (s *SubscriptionService) expirePartner(partnerID string, now time.Time) {
	partner, err := s.deps.Repos.Partners.GetByID(partnerID)
	if err != nil || partner == nil {
		return
	}
	if partner.Status != models.PartnerStatusActive {
		return
	}
	partner.Status = models.PartnerStatusExpired
	partner.UpdatedAt = now
	_ = s.deps.Repos.Partners.Update(partner)
}

// remainingQuota menghitung sisa kuota listing, atau -1 bila tanpa batas.
func remainingQuota(quota, used int) int {
	if quota < 0 {
		return -1
	}
	remaining := quota - used
	if remaining < 0 {
		return 0
	}
	return remaining
}
