// Package config memuat seluruh konfigurasi aplikasi dari environment variable.
//
// Nilai dibaca dari berkas .env (opsional, untuk pengembangan lokal) lalu
// ditimpa oleh environment variable proses. Setiap field memiliki default yang
// aman untuk pengembangan lokal, kecuali kredensial pihak ketiga yang wajib
// diisi manual (Firebase dan Midtrans).
package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config mengelompokkan konfigurasi per domain sesuai pembagian modul backend.
type Config struct {
	App       App       `envPrefix:"APP_"`
	HTTP      HTTP      `envPrefix:"HTTP_"`
	Database  Database  `envPrefix:"DB_"`
	Firebase  Firebase  `envPrefix:"FIREBASE_"`
	Auth      Auth      `envPrefix:"AUTH_"`
	Midtrans  Midtrans  `envPrefix:"MIDTRANS_"`
	Storage   Storage   `envPrefix:"STORAGE_"`
	CORS      CORS      `envPrefix:"CORS_"`
	RateLimit RateLimit `envPrefix:"RATE_LIMIT_"`
	Scheduler Scheduler `envPrefix:"SCHEDULER_"`
	Log       Log       `envPrefix:"LOG_"`
}

// App berisi identitas aplikasi dan alamat layanan eksternal yang dirujuk.
type App struct {
	Name                     string `env:"NAME" envDefault:"enmproperty-api"`
	Env                      string `env:"ENV" envDefault:"development"`
	Timezone                 string `env:"TIMEZONE" envDefault:"Asia/Jakarta"`
	FrontendURL              string `env:"FRONTEND_URL" envDefault:"http://localhost:3000"`
	WordPressURL             string `env:"WORDPRESS_URL" envDefault:""`
	WhatsAppAdmin            string `env:"WHATSAPP_ADMIN" envDefault:""`
	SubscriptionReminderDays int    `env:"SUBSCRIPTION_REMINDER_DAYS" envDefault:"7"`
}

// HTTP mengatur batas waktu, ukuran body, dan graceful shutdown HTTP server.
type HTTP struct {
	Port            string        `env:"PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" envDefault:"30s"`
	MaxBodyBytes    int64         `env:"MAX_BODY_BYTES" envDefault:"10485760"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
}

// Database mengatur koneksi dan pool PostgreSQL (pgxpool).
type Database struct {
	URL             string        `env:"URL" envDefault:"postgres://postgres:postgres@localhost:5432/enmproperty?sslmode=disable"`
	MaxConns        int32         `env:"MAX_CONNS" envDefault:"10"`
	MinConns        int32         `env:"MIN_CONNS" envDefault:"2"`
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME" envDefault:"1h"`
	MaxConnIdleTime time.Duration `env:"MAX_CONN_IDLE_TIME" envDefault:"30m"`
	MigrationsPath  string        `env:"MIGRATIONS_PATH" envDefault:"./migrations"`
	AutoMigrate     bool          `env:"AUTO_MIGRATE" envDefault:"false"`
	// Driver penyimpanan: memory (data contoh, tanpa PostgreSQL) atau postgres.
	Driver string `env:"DRIVER" envDefault:"memory"`
}

// Firebase memuat konfigurasi Firebase Admin SDK (auth dan storage bucket).
type Firebase struct {
	ProjectID   string `env:"PROJECT_ID"`
	DatabaseURL string `env:"DATABASE_URL"`
	// CredentialsFile diutamakan; bila kosong, CredentialsJSON dipakai.
	CredentialsFile string `env:"CREDENTIALS_FILE"`
	CredentialsJSON string `env:"CREDENTIALS_JSON"`
	StorageBucket   string `env:"STORAGE_BUCKET"`
}

// Auth mengatur jalur pengembangan verifikasi token.
//
// Mode dev hanya berlaku bila APP_ENV bukan production, sehingga token uji
// tetap tidak dapat dipakai pada deployment produksi.
type Auth struct {
	DevMode  bool   `env:"DEV_MODE" envDefault:"false"`
	DevToken string `env:"DEV_TOKEN" envDefault:"dev-token"`
}

// Midtrans memuat kredensial Midtrans Snap untuk checkout dan webhook.
// Nonaktif secara bawaan; aktifkan bersamaan dengan pengisian kuncinya.
type Midtrans struct {
	Enabled            bool   `env:"ENABLED" envDefault:"false"`
	ServerKey          string `env:"SERVER_KEY"`
	ClientKey          string `env:"CLIENT_KEY"`
	IsProduction       bool   `env:"IS_PRODUCTION" envDefault:"false"`
	SnapBaseURL        string `env:"SNAP_BASE_URL" envDefault:"https://app.sandbox.midtrans.com"`
	OrderExpiryMinutes int    `env:"ORDER_EXPIRY_MINUTES" envDefault:"60"`
}

// Storage mengatur penyimpanan foto properti beserta batasannya.
type Storage struct {
	Driver               string   `env:"DRIVER" envDefault:"local"`
	LocalDir             string   `env:"LOCAL_DIR" envDefault:"./storage"`
	PublicBaseURL        string   `env:"PUBLIC_BASE_URL" envDefault:"http://localhost:8080/uploads"`
	MaxUploadBytes       int64    `env:"MAX_UPLOAD_BYTES" envDefault:"5242880"`
	MaxPhotosPerProperty int      `env:"MAX_PHOTOS_PER_PROPERTY" envDefault:"10"`
	AllowedMimeTypes     []string `env:"ALLOWED_MIME_TYPES" envDefault:"image/jpeg,image/png,image/webp" envSeparator:","`
}

// CORS hanya mengizinkan origin aplikasi Next.js dan WordPress.
type CORS struct {
	AllowedOrigins   []string      `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000" envSeparator:","`
	AllowedMethods   []string      `env:"ALLOWED_METHODS" envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS" envSeparator:","`
	AllowedHeaders   []string      `env:"ALLOWED_HEADERS" envDefault:"Content-Type,Authorization,X-Request-Id" envSeparator:","`
	AllowCredentials bool          `env:"ALLOW_CREDENTIALS" envDefault:"true"`
	MaxAge           time.Duration `env:"MAX_AGE" envDefault:"12h"`
}

// RateLimit mengatur pembatasan request per IP pada route tertentu.
type RateLimit struct {
	Enabled          bool          `env:"ENABLED" envDefault:"true"`
	Store            string        `env:"STORE" envDefault:"memory"`
	RedisURL         string        `env:"REDIS_URL"`
	Requests         int           `env:"REQUESTS" envDefault:"60"`
	Window           time.Duration `env:"WINDOW" envDefault:"1m"`
	AuthRequests     int           `env:"AUTH_REQUESTS" envDefault:"20"`
	CheckoutRequests int           `env:"CHECKOUT_REQUESTS" envDefault:"10"`
	WebhookRequests  int           `env:"WEBHOOK_REQUESTS" envDefault:"120"`
}

// Scheduler mengatur job terjadwal (BE-16).
type Scheduler struct {
	Enabled                 bool   `env:"ENABLED" envDefault:"true"`
	ExpireSubscriptionsCron string `env:"EXPIRE_SUBSCRIPTIONS_CRON" envDefault:"0 1 * * *"`
	ExpireOrdersCron        string `env:"EXPIRE_ORDERS_CRON" envDefault:"*/15 * * * *"`
}

// Log mengatur format dan tingkat detail logging (zerolog).
type Log struct {
	Level  string `env:"LEVEL" envDefault:"info"`
	Format string `env:"FORMAT" envDefault:"json"`
}

// Load membaca konfigurasi dari berkas .env (bila ada) dan environment variable
// proses, lalu memvalidasi nilainya.
//
// Pada produksi berkas .env umumnya tidak ada karena nilai disuntikkan langsung
// oleh platform, sehingga kegagalan memuat berkas diabaikan.
func Load() (*Config, error) {
	_ = godotenv.Load(".env", ".env.local")

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("gagal memuat konfigurasi environment: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// IsProduction mengembalikan true bila aplikasi berjalan di environment produksi.
func (c *Config) IsProduction() bool { return c.App.Env == "production" }

// IsDevelopment mengembalikan true bila aplikasi berjalan di environment pengembangan.
func (c *Config) IsDevelopment() bool { return c.App.Env == "development" }

// UseDevAuth mengembalikan true bila token uji boleh dipakai. Selalu false di produksi.
func (c *Config) UseDevAuth() bool { return c.Auth.DevMode && !c.IsProduction() }

// UseMemoryStore mengembalikan true bila data disimpan di memori, bukan PostgreSQL.
func (c *Config) UseMemoryStore() bool { return c.Database.Driver != "postgres" }

// SimulatePayment mengembalikan true bila Midtrans belum dikonfigurasi sehingga
// checkout dijalankan dalam mode simulasi.
func (c *Config) SimulatePayment() bool {
	return !c.Midtrans.Enabled || c.Midtrans.ServerKey == ""
}

// Location mengembalikan lokasi waktu sesuai APP_TIMEZONE untuk format WIB.
func (c *Config) Location() (*time.Location, error) {
	return time.LoadLocation(c.App.Timezone)
}

func (c *Config) validate() error {
	var errs []error

	if !slices.Contains([]string{"development", "staging", "production"}, c.App.Env) {
		errs = append(errs, fmt.Errorf("APP_ENV harus development, staging, atau production (diterima %q)", c.App.Env))
	}
	if _, err := time.LoadLocation(c.App.Timezone); err != nil {
		errs = append(errs, fmt.Errorf("APP_TIMEZONE tidak dikenal (%q): %w", c.App.Timezone, err))
	}
	if strings.TrimSpace(c.Database.URL) == "" && !c.UseMemoryStore() {
		errs = append(errs, errors.New("DB_URL wajib diisi bila DB_DRIVER=postgres"))
	}
	if !slices.Contains([]string{"memory", "postgres"}, c.Database.Driver) {
		errs = append(errs, fmt.Errorf("DB_DRIVER harus memory atau postgres (diterima %q)", c.Database.Driver))
	}
	if len(c.CORS.AllowedOrigins) == 0 {
		errs = append(errs, errors.New("CORS_ALLOWED_ORIGINS wajib diisi minimal satu origin"))
	}
	if c.Auth.DevMode && c.IsProduction() {
		errs = append(errs, errors.New("AUTH_DEV_MODE tidak boleh aktif saat APP_ENV=production"))
	}
	if c.Auth.DevMode && strings.TrimSpace(c.Auth.DevToken) == "" {
		errs = append(errs, errors.New("AUTH_DEV_TOKEN wajib diisi bila AUTH_DEV_MODE=true"))
	}

	switch c.Storage.Driver {
	case "local":
		if c.Storage.LocalDir == "" {
			errs = append(errs, errors.New("STORAGE_LOCAL_DIR wajib diisi bila STORAGE_DRIVER=local"))
		}
	case "gcs":
		if c.Firebase.StorageBucket == "" {
			errs = append(errs, errors.New("FIREBASE_STORAGE_BUCKET wajib diisi bila STORAGE_DRIVER=gcs"))
		}
	default:
		errs = append(errs, fmt.Errorf("STORAGE_DRIVER harus local atau gcs (diterima %q)", c.Storage.Driver))
	}

	if c.Midtrans.Enabled {
		if c.Midtrans.ServerKey == "" {
			errs = append(errs, errors.New("MIDTRANS_SERVER_KEY wajib diisi bila MIDTRANS_ENABLED=true"))
		}
		if c.Midtrans.ClientKey == "" {
			errs = append(errs, errors.New("MIDTRANS_CLIENT_KEY wajib diisi bila MIDTRANS_ENABLED=true"))
		}
	}

	switch c.RateLimit.Store {
	case "memory":
	case "redis":
		if c.RateLimit.RedisURL == "" {
			errs = append(errs, errors.New("RATE_LIMIT_REDIS_URL wajib diisi bila RATE_LIMIT_STORE=redis"))
		}
	default:
		errs = append(errs, fmt.Errorf("RATE_LIMIT_STORE harus memory atau redis (diterima %q)", c.RateLimit.Store))
	}

	if !slices.Contains([]string{"json", "console"}, c.Log.Format) {
		errs = append(errs, fmt.Errorf("LOG_FORMAT harus json atau console (diterima %q)", c.Log.Format))
	}
	if !slices.Contains([]string{"debug", "info", "warn", "error"}, strings.ToLower(c.Log.Level)) {
		errs = append(errs, fmt.Errorf("LOG_LEVEL harus debug, info, warn, atau error (diterima %q)", c.Log.Level))
	}

	return errors.Join(errs...)
}
