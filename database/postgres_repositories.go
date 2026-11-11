package database

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresRepositories membuat kumpulan repository berbasis PostgreSQL
// (pgxpool) yang setara dengan implementasi in-memory. Seluruh struct
// repository menerima pool melalui antarmuka pgxPool sehingga tetap mudah
// diuji, sedangkan pembuat ini menerima *pgxpool.Pool konkret dari main.
func NewPostgresRepositories(pool *pgxpool.Pool) Repositories {
	return Repositories{
		Users:         &postgresUserRepository{pool: pool},
		Partners:      &postgresPartnerRepository{pool: pool},
		Plans:         &postgresPlanRepository{pool: pool},
		Categories:    &postgresCategoryRepository{pool: pool},
		Facilities:    &postgresFacilityRepository{pool: pool},
		Agreements:    &postgresAgreementRepository{pool: pool},
		Subscriptions: &postgresSubscriptionRepository{pool: pool},
		Orders:        &postgresOrderRepository{pool: pool},
		Properties:    &postgresPropertyRepository{pool: pool},
		Favorites:     &postgresFavoriteRepository{pool: pool},
		Analytics:     &postgresAnalyticsRepository{pool: pool},
		Settings:      &postgresSettingRepository{pool: pool},
		Landings:      &postgresLandingRepository{pool: pool},
		Audits:        &postgresAuditRepository{pool: pool},
	}
}
