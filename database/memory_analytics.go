package database

import (
	"enmproperty/backend/models"
)

type memoryAnalyticsRepository struct{ store *Store }

// AddView mencatat tayangan sekaligus menaikkan penghitung pada properti.
func (r *memoryAnalyticsRepository) AddView(view *models.PropertyView) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.views = append(r.store.views, copyView(view))
	if property, exists := r.store.properties[view.PropertyID]; exists {
		property.Views++
	}
	return nil
}

// AddContact mencatat lead WhatsApp sekaligus menaikkan penghitung pada properti.
func (r *memoryAnalyticsRepository) AddContact(contact *models.PropertyContact) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.contacts = append(r.store.contacts, copyContact(contact))
	if property, exists := r.store.properties[contact.PropertyID]; exists {
		property.ContactCount++
	}
	return nil
}

func (r *memoryAnalyticsRepository) Views() []*models.PropertyView {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	views := make([]*models.PropertyView, 0, len(r.store.views))
	for _, view := range r.store.views {
		views = append(views, copyView(view))
	}
	return views
}

func (r *memoryAnalyticsRepository) Contacts() []*models.PropertyContact {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	contacts := make([]*models.PropertyContact, 0, len(r.store.contacts))
	for _, contact := range r.store.contacts {
		contacts = append(contacts, copyContact(contact))
	}
	return contacts
}
