package database

import (
	"strconv"
	"time"

	"enmproperty/backend/models"
)

// ID tetap untuk data contoh agar frontend dapat dikembangkan tanpa backend
// menyalakan basis data sungguhan.
const (
	SeedUserSuperadmin = "usr-superadmin"
	SeedUserAdmin      = "usr-admin"
	SeedUserMitraBali  = "usr-mitra-bali"
	SeedUserMitraJogja = "usr-mitra-jogja"
	SeedUserCustomer   = "usr-customer"

	SeedPartnerBali  = "ptr-bali"
	SeedPartnerJogja = "ptr-jogja"

	SeedPlanFree    = "pln-free"
	SeedPlanStandar = "pln-standar"
	SeedPlanPremium = "pln-premium"

	SeedAgreement = "agr-2026-01"
)

// Seed mengisi Store dengan data contoh yang mencerminkan alur nyata: satu
// superadmin, satu admin, dua mitra dengan langganan, dan beberapa properti.
// Panggilan berulang akan menimpa data yang sama sehingga aman dipanggil saat
// pengembangan.
func Seed(store *Store) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	store.mu.Lock()
	defer store.mu.Unlock()

	seedUsers(store, today)
	seedPlans(store, today)
	seedAgreement(store, today)
	seedPartners(store, today)
	seedSubscriptions(store, today)
	seedOrders(store, today)
	seedProperties(store, today)
	seedSettings(store, now)
}

func seedUsers(store *Store, today time.Time) {
	lastLogin := today.Add(3 * time.Hour)
	users := []*models.User{
		{
			ID: SeedUserSuperadmin, FirebaseUID: "dev-superadmin", Email: "superadmin@enmproperty.id",
			Name: "Bagas Superadmin", Phone: "+6281100000001", Role: models.RoleSuperadmin,
			Status: models.UserStatusActive, EmailVerified: true, LastLoginAt: &lastLogin,
			CreatedAt: today.AddDate(0, -8, 0), UpdatedAt: today,
		},
		{
			ID: SeedUserAdmin, FirebaseUID: "dev-admin", Email: "admin@enmproperty.id",
			Name: "Rani Admin", Phone: "+6281100000002", Role: models.RoleAdmin,
			Status: models.UserStatusActive, EmailVerified: true, LastLoginAt: &lastLogin,
			CreatedAt: today.AddDate(0, -8, 0), UpdatedAt: today,
		},
		{
			ID: SeedUserMitraBali, FirebaseUID: "dev-mitra-bali", Email: "mitra.bali@enmproperty.id",
			Name: "Wayan Sudira", Phone: "+6281200000001", Role: models.RoleMitra,
			Status: models.UserStatusActive, EmailVerified: true, LastLoginAt: &lastLogin,
			CreatedAt: today.AddDate(0, -6, 0), UpdatedAt: today,
		},
		{
			ID: SeedUserMitraJogja, FirebaseUID: "dev-mitra-jogja", Email: "mitra.jogja@enmproperty.id",
			Name: "Sri Handayani", Phone: "+6281200000002", Role: models.RoleMitra,
			Status: models.UserStatusActive, EmailVerified: true, LastLoginAt: &lastLogin,
			CreatedAt: today.AddDate(0, -5, 0), UpdatedAt: today,
		},
		{
			ID: SeedUserCustomer, FirebaseUID: "dev-customer", Email: "customer@example.com",
			Name: "Dewi Lestari", Phone: "+6281300000001", Role: models.RoleCustomer,
			Status: models.UserStatusActive, EmailVerified: true, LastLoginAt: &lastLogin,
			CreatedAt: today.AddDate(0, -2, 0), UpdatedAt: today,
		},
	}
	for _, user := range users {
		store.users[user.ID] = copyUser(user)
	}
}

func seedPlans(store *Store, today time.Time) {
	plans := []*models.Plan{
		{
			ID: SeedPlanFree, Slug: "gratis", Name: "Gratis", Description: "Paket uji coba untuk mitra baru.",
			Price: 0, DurationDays: 30, ListingQuota: 3, IsFree: true, IsActive: true, SortOrder: 1,
			Features:  []string{"3 listing aktif", "Masa aktif 30 hari", "Dukungan email"},
			CreatedAt: today.AddDate(0, -8, 0), UpdatedAt: today,
		},
		{
			ID: SeedPlanStandar, Slug: "standar", Name: "Standar", Description: "Cocok untuk agen properti perorangan.",
			Price: 750000, DurationDays: 90, ListingQuota: 15, IsActive: true, SortOrder: 2,
			Features:  []string{"15 listing aktif", "Masa aktif 90 hari", "Statistik tayangan", "Dukungan WhatsApp"},
			CreatedAt: today.AddDate(0, -8, 0), UpdatedAt: today,
		},
		{
			ID: SeedPlanPremium, Slug: "premium", Name: "Premium", Description: "Untuk agensi dengan banyak listing.",
			Price: 1500000, DurationDays: 90, ListingQuota: 25, IsActive: true, SortOrder: 3,
			Features:  []string{"25 listing aktif", "Masa aktif 90 hari", "Prioritas review", "Statistik lanjutan", "Dukungan prioritas"},
			CreatedAt: today.AddDate(0, -8, 0), UpdatedAt: today,
		},
	}
	for _, plan := range plans {
		store.plans[plan.ID] = copyPlan(plan)
	}
}

func seedAgreement(store *Store, today time.Time) {
	agreement := &models.Agreement{
		ID: SeedAgreement, Version: "2026.01", Title: "Perjanjian Kemitraan ENM Property",
		Content:  "Dengan mendaftar sebagai mitra, Anda menyetujui ketentuan pemuatan listing, kebenaran data properti, dan kebijakan moderasi platform.",
		IsActive: true, EffectiveAt: today.AddDate(0, -6, 0), CreatedBy: SeedUserSuperadmin,
		CreatedAt: today.AddDate(0, -6, 0),
	}
	store.agreements[agreement.ID] = copyAgreement(agreement)
}

func seedPartners(store *Store, today time.Time) {
	joinedBali := today.AddDate(0, -6, 0)
	joinedJogja := today.AddDate(0, -5, 0)
	partners := []*models.Partner{
		{
			ID: SeedPartnerBali, UserID: SeedUserMitraBali, BusinessName: "Villa Bali Sejahtera",
			PICName: "Wayan Sudira", Phone: "+6281200000001", Email: "mitra.bali@enmproperty.id",
			Address: "Jl. Sunset Road No. 88", City: "Badung", Province: "Bali",
			Status: models.PartnerStatusActive, JoinedAt: &joinedBali,
			CreatedAt: today.AddDate(0, -6, 0), UpdatedAt: today,
		},
		{
			ID: SeedPartnerJogja, UserID: SeedUserMitraJogja, BusinessName: "Griya Properti Jogja",
			PICName: "Sri Handayani", Phone: "+6281200000002", Email: "mitra.jogja@enmproperty.id",
			Address: "Jl. Kaliurang KM 5", City: "Sleman", Province: "DI Yogyakarta",
			Status: models.PartnerStatusActive, JoinedAt: &joinedJogja,
			CreatedAt: today.AddDate(0, -5, 0), UpdatedAt: today,
		},
	}
	for _, partner := range partners {
		store.partners[partner.ID] = copyPartner(partner)
	}
}

func seedSubscriptions(store *Store, today time.Time) {
	subscriptions := []*models.Subscription{
		{
			ID: "sub-bali-premium", PartnerID: SeedPartnerBali, PlanID: SeedPlanPremium, PlanName: "Premium",
			OrderID: "ord-bali-premium", StartsAt: today.AddDate(0, -2, 0), EndsAt: today.AddDate(0, 1, 0),
			Status: models.SubscriptionStatusActive, CreatedAt: today.AddDate(0, -2, 0), UpdatedAt: today,
		},
		{
			ID: "sub-jogja-standar", PartnerID: SeedPartnerJogja, PlanID: SeedPlanStandar, PlanName: "Standar",
			OrderID: "ord-jogja-standar", StartsAt: today.AddDate(0, -1, 0), EndsAt: today.AddDate(0, 2, 0),
			Status: models.SubscriptionStatusActive, CreatedAt: today.AddDate(0, -1, 0), UpdatedAt: today,
		},
	}
	for _, subscription := range subscriptions {
		store.subscriptions[subscription.ID] = copySubscription(subscription)
	}
}

func seedOrders(store *Store, today time.Time) {
	paidBali := today.AddDate(0, -2, 0)
	paidJogja := today.AddDate(0, -1, 0)
	orders := []*models.Order{
		{
			ID: "ord-bali-premium", Code: "INV-2026-0001", UserID: SeedUserMitraBali, PartnerID: SeedPartnerBali,
			PartnerName: "Villa Bali Sejahtera", PlanID: SeedPlanPremium, PlanName: "Premium",
			PlanPrice: 1500000, PlanDurationDays: 90, PlanListingQuota: 25, Amount: 1500000, Kind: models.OrderKindNew,
			Status: models.OrderStatusPaid, Provider: models.PaymentProviderSimulated, PaidAt: &paidBali,
			ExpiresAt: paidBali.Add(24 * time.Hour), CreatedAt: paidBali, UpdatedAt: paidBali,
		},
		{
			ID: "ord-jogja-standar", Code: "INV-2026-0002", UserID: SeedUserMitraJogja, PartnerID: SeedPartnerJogja,
			PartnerName: "Griya Properti Jogja", PlanID: SeedPlanStandar, PlanName: "Standar",
			PlanPrice: 750000, PlanDurationDays: 90, PlanListingQuota: 15, Amount: 750000, Kind: models.OrderKindNew,
			Status: models.OrderStatusPaid, Provider: models.PaymentProviderSimulated, PaidAt: &paidJogja,
			ExpiresAt: paidJogja.Add(24 * time.Hour), CreatedAt: paidJogja, UpdatedAt: paidJogja,
		},
	}
	for _, order := range orders {
		store.orders[order.ID] = copyOrder(order)
	}
}

func seedProperties(store *Store, today time.Time) {
	type seedPhoto struct {
		url     string
		primary bool
	}
	type seedProperty struct {
		property *models.Property
		photos   []seedPhoto
	}

	seeds := []seedProperty{
		{
			property: &models.Property{
				ID: "prp-villa-seminyak", PartnerID: SeedPartnerBali, PartnerName: "Villa Bali Sejahtera",
				Slug: "villa-seminyak-modern", Title: "Villa Seminyak Modern",
				Description: "Villa dua lantai dengan kolam renang pribadi, lima menit dari Pantai Seminyak.",
				ListingType: models.ListingTypeSale, PropertyType: models.PropertyTypeVilla,
				Address: "Jl. Kayu Aya No. 12", City: "Badung", Province: "Bali", District: "Seminyak",
				Latitude: -8.6833, Longitude: 115.1667, Price: 4500000000,
				LandArea: 300, BuildingArea: 240, Bedrooms: 4, Bathrooms: 3,
				Facilities: []string{"Kolam renang", "Carport", "Keamanan 24 jam", "Taman"},
				UnitsTotal: 1, UnitsAvailable: 1, Status: models.PropertyStatusActive, Views: 862,
				PublishedAt: timePtr(today.AddDate(0, -2, 0)),
				CreatedAt:   today.AddDate(0, -2, 0), UpdatedAt: today,
			},
			photos: []seedPhoto{
				{url: "https://images.unsplash.com/photo-1613490493576-7fde63acd811", primary: true},
				{url: "https://images.unsplash.com/photo-1600585154340-be6161a56a0c"},
			},
		},
		{
			property: &models.Property{
				ID: "prp-cluster-sentul", PartnerID: SeedPartnerJogja, PartnerName: "Griya Properti Jogja",
				Slug: "rumah-cluster-sentul", Title: "Rumah Cluster Sentul",
				Description: "Rumah modern minimalis dalam cluster dengan akses tol dan area hijau.",
				ListingType: models.ListingTypeSale, PropertyType: models.PropertyTypeHouse,
				Address: "Cluster Sentul City Blok B2", City: "Bogor", Province: "Jawa Barat", District: "Sentul",
				Latitude: -6.5670, Longitude: 106.8547, Price: 1250000000,
				LandArea: 120, BuildingArea: 90, Bedrooms: 3, Bathrooms: 2,
				Facilities: []string{"Keamanan 24 jam", "Taman", "Playground"},
				UnitsTotal: 1, UnitsAvailable: 1, Status: models.PropertyStatusPendingReview,
				CreatedAt: today.AddDate(0, 0, -6), UpdatedAt: today,
			},
			photos: []seedPhoto{{url: "https://images.unsplash.com/photo-1568605114967-8130f3a36994", primary: true}},
		},
		{
			property: &models.Property{
				ID: "prp-ruko-solo", PartnerID: SeedPartnerJogja, PartnerName: "Griya Properti Jogja",
				Slug: "ruko-pusat-kota-solo", Title: "Ruko Pusat Kota Solo",
				Description: "Ruko tiga lantai di kawasan bisnis strategis dengan parkir luas.",
				ListingType: models.ListingTypeRent, PropertyType: models.PropertyTypeRuko, PricePeriod: models.PricePeriodYearly,
				Address: "Jl. Slamet Riyadi No. 210", City: "Surakarta", Province: "Jawa Tengah", District: "Laweyan",
				Latitude: -7.5666, Longitude: 110.8167, Price: 180000000,
				LandArea: 100, BuildingArea: 240, Bedrooms: 0, Bathrooms: 3,
				Facilities: []string{"Parkir luas", "Listrik 3 fase"},
				UnitsTotal: 2, UnitsAvailable: 1, Status: models.PropertyStatusActive, Views: 415,
				PublishedAt: timePtr(today.AddDate(0, -3, 0)),
				CreatedAt:   today.AddDate(0, -3, 0), UpdatedAt: today,
			},
			photos: []seedPhoto{{url: "https://images.unsplash.com/photo-1524758631624-e2822e304c36", primary: true}},
		},
		{
			property: &models.Property{
				ID: "prp-tanah-ungaran", PartnerID: SeedPartnerJogja, PartnerName: "Griya Properti Jogja",
				Slug: "tanah-kavling-ungaran", Title: "Tanah Kavling Ungaran",
				Description: "Kavling tanah siap bangun dengan akses jalan aspal dan pemandangan pegunungan.",
				ListingType: models.ListingTypeSale, PropertyType: models.PropertyTypeLand,
				Address: "Jl. Raya Ungaran KM 3", City: "Semarang", Province: "Jawa Tengah", District: "Ungaran",
				Latitude: -7.1387, Longitude: 110.4031, Price: 650000000,
				LandArea: 200, BuildingArea: 0, Bedrooms: 0, Bathrooms: 0,
				Facilities: []string{"Akses jalan aspal", "Sertifikat SHM"},
				UnitsTotal: 4, UnitsAvailable: 2, Status: models.PropertyStatusExpired, Views: 96,
				CreatedAt: today.AddDate(0, -4, 0), UpdatedAt: today,
			},
			photos: []seedPhoto{{url: "https://images.unsplash.com/photo-1500382017468-9049fed747ef", primary: true}},
		},
	}

	for _, seed := range seeds {
		store.properties[seed.property.ID] = copyProperty(seed.property)
		for i, photo := range seed.photos {
			id := seed.property.ID + "-photo-" + strconv.Itoa(i+1)
			stored := &models.PropertyPhoto{
				ID: id, PropertyID: seed.property.ID, URL: photo.url, StoragePath: "seed/" + id + ".jpg",
				MimeType: "image/jpeg", SortOrder: i, IsPrimary: photo.primary, CreatedAt: seed.property.CreatedAt,
			}
			store.photos[id] = stored
		}
	}
}

func seedSettings(store *Store, now time.Time) {
	settings := []*models.Setting{
		{Key: models.SettingSiteName, Value: "ENM Property", Description: "Nama situs yang tampil di header dan metadata.", UpdatedAt: now},
		{Key: models.SettingAdminWhatsApp, Value: "+6281100000000", Description: "Nomor WhatsApp admin untuk inquiry mitra.", UpdatedAt: now},
		{Key: models.SettingSubscriptionAlertDay, Value: "7", Description: "Ambang hari peringatan sebelum langganan berakhir.", UpdatedAt: now},
		{Key: models.SettingContactFallback, Value: "+6281100000000", Description: "Nomor fallback bila mitra belum mengisi kontak.", UpdatedAt: now},
		{Key: models.SettingAgreementEnabled, Value: "true", Description: "Wajib menyetujui perjanjian sebelum mengunggah listing.", UpdatedAt: now},
	}
	for _, setting := range settings {
		store.settings[setting.Key] = copySetting(setting)
	}
}

func timePtr(t time.Time) *time.Time { return &t }
