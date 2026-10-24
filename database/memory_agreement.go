package database

import (
	"enmproperty/backend/models"
)

type memoryAgreementRepository struct{ store *Store }

func (r *memoryAgreementRepository) Create(agreement *models.Agreement) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.agreements[agreement.ID] = copyAgreement(agreement)
	return nil
}

func (r *memoryAgreementRepository) Update(agreement *models.Agreement) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.agreements[agreement.ID]; !exists {
		return ErrNotFound
	}
	r.store.agreements[agreement.ID] = copyAgreement(agreement)
	return nil
}

func (r *memoryAgreementRepository) GetByID(id string) (*models.Agreement, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	agreement, exists := r.store.agreements[id]
	if !exists {
		return nil, ErrNotFound
	}
	return copyAgreement(agreement), nil
}

func (r *memoryAgreementRepository) GetActive() (*models.Agreement, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	var newest *models.Agreement
	for _, agreement := range r.store.agreements {
		if !agreement.IsActive {
			continue
		}
		if newest == nil || agreement.EffectiveAt.After(newest.EffectiveAt) {
			newest = agreement
		}
	}
	if newest == nil {
		return nil, ErrNotFound
	}
	return copyAgreement(newest), nil
}

func (r *memoryAgreementRepository) List(q ListQuery) ([]*models.Agreement, int, error) {
	items := make([]*models.Agreement, 0)
	r.store.mu.RLock()
	for _, agreement := range r.store.agreements {
		items = append(items, copyAgreement(agreement))
	}
	r.store.mu.RUnlock()

	SortByFunc(items, func(a, b *models.Agreement) bool {
		return a.EffectiveAt.Before(b.EffectiveAt)
	}, true)

	page, total := Paginate(items, q)
	return page, total, nil
}

func (r *memoryAgreementRepository) SaveAcceptance(acceptance *models.AgreementAcceptance) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.acceptances[acceptance.ID] = copyAcceptance(acceptance)
	return nil
}

func (r *memoryAgreementRepository) GetAcceptance(partnerID, agreementID string) (*models.AgreementAcceptance, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, acceptance := range r.store.acceptances {
		if acceptance.PartnerID == partnerID && acceptance.AgreementID == agreementID {
			return copyAcceptance(acceptance), nil
		}
	}
	return nil, ErrNotFound
}

func (r *memoryAgreementRepository) ListAcceptances(q ListQuery) ([]*models.AgreementAcceptance, int, error) {
	items := make([]*models.AgreementAcceptance, 0)
	r.store.mu.RLock()
	for _, acceptance := range r.store.acceptances {
		items = append(items, copyAcceptance(acceptance))
	}
	r.store.mu.RUnlock()

	filtered := make([]*models.AgreementAcceptance, 0, len(items))
	for _, acceptance := range items {
		if partnerID := q.Filter("partner_id"); partnerID != "" && acceptance.PartnerID != partnerID {
			continue
		}
		if agreementID := q.Filter("agreement_id"); agreementID != "" && acceptance.AgreementID != agreementID {
			continue
		}
		filtered = append(filtered, acceptance)
	}

	SortByFunc(filtered, func(a, b *models.AgreementAcceptance) bool {
		return a.AcceptedAt.Before(b.AcceptedAt)
	}, true)

	page, total := Paginate(filtered, q)
	return page, total, nil
}
