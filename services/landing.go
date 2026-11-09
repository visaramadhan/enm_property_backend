package services

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
)

// LandingService mengelola konten landing page (CMS Landing Page) dan memicu
// revalidasi ISR pada frontend setelah perubahan dipublikasikan.
type LandingService struct {
	deps Deps

	mu     sync.RWMutex
	cached *dto.LandingContentView
}

// Content mengembalikan konten landing page siap pakai. Hasilnya di-cache
// sampai ada perubahan agar endpoint publik tidak menyentuh repository pada
// setiap permintaan.
func (s *LandingService) Content() dto.LandingContentView {
	if cached := s.loadCache(); cached != nil {
		return *cached
	}

	view := dto.NewLandingContentView(s.deps.Repos.Landings.All())
	s.storeCache(&view)
	return view
}

// Update menyimpan seluruh konten landing page lalu menyegarkan cache.
func (s *LandingService) Update(actorID string, req dto.UpdateLandingRequest) (dto.LandingContentView, error) {
	now := timeNow()
	values := map[string]string{
		models.LandingHeroTitle:       strings.TrimSpace(req.HeroTitle),
		models.LandingHeroSubtitle:    strings.TrimSpace(req.HeroSubtitle),
		models.LandingHeroCTAText:     strings.TrimSpace(req.HeroCTAText),
		models.LandingHeroImageURL:    strings.TrimSpace(req.HeroImageURL),
		models.LandingMetaTitle:       strings.TrimSpace(req.MetaTitle),
		models.LandingMetaDescription: strings.TrimSpace(req.MetaDescription),
		models.LandingMetaOGImage:     strings.TrimSpace(req.MetaOGImage),
		models.LandingHighlightAreas:  dto.EncodeStringList(cleanAreas(req.HighlightAreas)),
		models.LandingCTATitle:        strings.TrimSpace(req.CTATitle),
		models.LandingCTADescription:  strings.TrimSpace(req.CTADescription),
	}

	for key, value := range values {
		item := &models.LandingContent{
			Key:       key,
			Value:     value,
			UpdatedBy: actorID,
			UpdatedAt: now,
		}
		if err := s.deps.Repos.Landings.Set(item); err != nil {
			return dto.LandingContentView{}, httpx.Internal(err)
		}
	}

	return s.refresh(), nil
}

// Revalidate memicu revalidasi ISR halaman landing page pada frontend melalui
// webhook /api/revalidate. Dipanggil tombol "Publish Perubahan Landing Page".
func (s *LandingService) Revalidate(ctx context.Context) error {
	token := strings.TrimSpace(s.deps.Config.App.RevalidationToken)
	base := strings.TrimRight(strings.TrimSpace(s.deps.Config.App.FrontendURL), "/")
	if token == "" || base == "" {
		return httpx.Unprocessable("Revalidasi belum dikonfigurasi pada server")
	}

	endpoint := base + "/api/revalidate?secret=" + url.QueryEscape(token) + "&path=/"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return httpx.Internal(err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return httpx.Internal(err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusMultipleChoices {
		return httpx.Unprocessable("Revalidasi halaman gagal (status " + response.Status + ")")
	}
	return nil
}

// refresh membaca ulang konten dari repository dan memperbarui cache.
func (s *LandingService) refresh() dto.LandingContentView {
	view := dto.NewLandingContentView(s.deps.Repos.Landings.All())
	s.storeCache(&view)
	return view
}

// cleanAreas membuang area kosong dan duplikat tanpa mengubah urutan.
func cleanAreas(areas []string) []string {
	hasil := make([]string, 0, len(areas))
	terlihat := make(map[string]struct{}, len(areas))
	for _, area := range areas {
		bersih := strings.TrimSpace(area)
		if bersih == "" {
			continue
		}
		if _, ada := terlihat[bersih]; ada {
			continue
		}
		terlihat[bersih] = struct{}{}
		hasil = append(hasil, bersih)
	}
	return hasil
}

// loadCache membaca konten landing dari cache.
func (s *LandingService) loadCache() *dto.LandingContentView {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cached
}

// storeCache menyimpan konten landing ke cache.
func (s *LandingService) storeCache(view *dto.LandingContentView) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cached = view
}
