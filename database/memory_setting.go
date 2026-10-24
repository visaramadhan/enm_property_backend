package database

import (
	"sort"

	"enmproperty/backend/models"
)

type memorySettingRepository struct{ store *Store }

func (r *memorySettingRepository) All() []*models.Setting {
	r.store.mu.RLock()
	settings := make([]*models.Setting, 0, len(r.store.settings))
	for _, setting := range r.store.settings {
		settings = append(settings, copySetting(setting))
	}
	r.store.mu.RUnlock()

	sort.SliceStable(settings, func(i, j int) bool { return settings[i].Key < settings[j].Key })
	return settings
}

func (r *memorySettingRepository) Get(key string) (*models.Setting, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	setting, exists := r.store.settings[key]
	if !exists {
		return nil, ErrNotFound
	}
	return copySetting(setting), nil
}

func (r *memorySettingRepository) Set(setting *models.Setting) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if existing, exists := r.store.settings[setting.Key]; exists {
		stored := copySetting(setting)
		if stored.Description == "" {
			stored.Description = existing.Description
		}
		stored.IsSecret = existing.IsSecret || setting.IsSecret
		r.store.settings[setting.Key] = stored
		return nil
	}
	r.store.settings[setting.Key] = copySetting(setting)
	return nil
}
