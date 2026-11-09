// Package router merakit seluruh route HTTP beserta rantai middleware-nya.
//
// Handler tidak memuat logika bisnis; router hanya memetakan method dan path ke
// handler, serta memasang middleware sesuai urutan yang ditetapkan pada
// spesifikasi backend (autentikasi, peran, status mitra, paket aktif, audit).
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"enmproperty/backend/config"
	"enmproperty/backend/database"
	"enmproperty/backend/handler"
	"enmproperty/backend/middleware"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/firebase"
	"enmproperty/backend/platform/midtrans"
	"enmproperty/backend/services"
)

// Deps mengumpulkan dependensi yang dibutuhkan untuk merakit router.
type Deps struct {
	Config   *config.Config
	Services *services.Services
	Repos    database.Repositories
	Firebase *firebase.Client
	Midtrans *midtrans.Client
	Logger   zerolog.Logger
}

// New membentuk *gin.Engine lengkap dengan middleware lintas potong dan seluruh
// route per modul di bawah prefiks /api/v1.
func New(deps Deps) *gin.Engine {
	cfg := deps.Config
	svc := deps.Services

	engine := gin.New()
	engine.Use(
		middleware.Recovery(deps.Logger),
		middleware.ErrorHandler(),
		middleware.RequestID(),
		middleware.AccessLogger(deps.Logger),
		middleware.SecureHeaders(),
		middleware.CORS(cfg.CORS),
	)

	// Handler per modul.
	healthHandler := handler.NewHealthHandler(svc)
	authHandler := handler.NewAuthHandler(svc)
	meHandler := handler.NewMeHandler(svc)
	planHandler := handler.NewPlanHandler(svc)
	categoryHandler := handler.NewCategoryHandler(svc)
	catalogHandler := handler.NewCatalogHandler(svc)
	favoriteHandler := handler.NewFavoriteHandler(svc)
	agreementHandler := handler.NewAgreementHandler(svc)
	partnerHandler := handler.NewPartnerHandler(svc)
	checkoutHandler := handler.NewCheckoutHandler(svc)
	webhookHandler := handler.NewMidtransWebhookHandler(svc)
	invoiceHandler := handler.NewInvoiceHandler(svc)
	vendorDashboardHandler := handler.NewVendorDashboardHandler(svc)
	vendorSubscriptionHandler := handler.NewVendorSubscriptionHandler(svc)
	vendorPropertyHandler := handler.NewVendorPropertyHandler(svc)
	vendorMediaHandler := handler.NewVendorMediaHandler(svc)
	adminDashboardHandler := handler.NewAdminDashboardHandler(svc)
	adminPartnerHandler := handler.NewAdminPartnerHandler(svc)
	adminPropertyHandler := handler.NewAdminPropertyHandler(svc)
	adminMediaHandler := handler.NewAdminMediaHandler(svc)
	adminReportHandler := handler.NewAdminReportHandler(svc)
	adminUserHandler := handler.NewAdminUserHandler(svc)
	adminPlanHandler := handler.NewAdminPlanHandler(svc)
	adminCategoryHandler := handler.NewAdminCategoryHandler(svc)
	adminAgreementHandler := handler.NewAdminAgreementHandler(svc)
	adminSettingHandler := handler.NewAdminSettingHandler(svc)
	siteSettingHandler := handler.NewSiteSettingHandler(svc)
	landingHandler := handler.NewLandingHandler(svc)
	adminAuditHandler := handler.NewAdminAuditHandler(svc)
	ownerHandler := handler.NewOwnerHandler(svc)

	// Middleware yang dipakai berulang.
	authRequired := middleware.FirebaseAuth(deps.Firebase, deps.Repos.Users)
	activePartner := middleware.RequireActivePartner(deps.Repos.Partners)
	activePackage := middleware.RequireActivePackage(deps.Repos.Subscriptions)
	auditHook := middleware.AuditHook(deps.Repos.Audits)
	generalLimit := middleware.RateLimit(cfg.RateLimit, cfg.RateLimit.Requests)
	bodyLimit := middleware.BodyLimit(cfg.HTTP.MaxBodyBytes)
	requestTimeout := middleware.Timeout(cfg.HTTP.RequestTimeout)

	// Route publik dan terautentikasi memakai batas body standar.
	//
	// Kompresi (gzip) sengaja dipasang SETELAH Timeout pada group ini, bukan di
	// engine-level. `gin-contrib/timeout` membungkus ResponseWriter dengan buffer
	// lalu men-set `Size()` menjadi -1 saat flush; bila gzip berada di luar, defer
	// `gin-contrib/gzip` membaca `Size() < 0` dan membuang seluruh body respons
	// (hanya menyisakan header gzip 10 byte) sehingga klien menerima body kosong.
	// Dengan gzip di dalam Timeout, gzip menulis ke buffer Timeout dan ukuran
	// respons tetap valid saat defer-nya berjalan.
	api := engine.Group("/api/v1", generalLimit, requestTimeout, bodyLimit, middleware.Compress())
	// Route unggah berkas tidak memakai batas body standar karena ukurannya
	// dibatasi oleh UploadGuard dengan kuota tersendiri.
	uploads := engine.Group("/api/v1", generalLimit, requestTimeout, middleware.Compress())

	// --- Endpoint publik ---------------------------------------------------
	api.GET("/health", healthHandler.Health)
	api.GET("/health/ready", healthHandler.Ready)

	api.POST("/auth/session",
		middleware.RateLimit(cfg.RateLimit, cfg.RateLimit.AuthRequests),
		middleware.Authenticate(deps.Firebase),
		authHandler.Session,
	)

	api.GET("/packages", planHandler.List)
	api.GET("/packages/:slug", planHandler.Detail)

	// Kategori properti untuk filter katalog publik.
	api.GET("/categories", categoryHandler.List)

	api.GET("/properties", catalogHandler.Search)
	api.GET("/properties/:slug", catalogHandler.Detail)
	// Form kontak publik pada halaman detail properti (lead kanal "form").
	// Memakai autentikasi opsional: pengunjung anonim tetap dilayani, sedangkan
	// pengunjung yang sudah login ikut tercatat pada lead.
	api.POST("/properties/:slug/contact-form",
		middleware.OptionalAuth(deps.Firebase, deps.Repos.Users),
		catalogHandler.ContactForm,
	)

	// Profil pengelola (mitra) beserta katalog publiknya.
	api.GET("/partners/:id", catalogHandler.Partner)

	api.GET("/agreements/active", agreementHandler.Active)

	// Pengaturan tampilan (nama situs, logo, gambar area masuk) untuk halaman publik.
	api.GET("/settings", siteSettingHandler.Get)

	// Konten landing page yang dirender ISR oleh frontend.
	api.GET("/landing", landingHandler.Get)

	api.POST("/payments/midtrans/notification",
		middleware.RateLimit(cfg.RateLimit, cfg.RateLimit.WebhookRequests),
		middleware.MidtransSignature(deps.Midtrans),
		webhookHandler.Notification,
	)

	// --- Endpoint login (semua peran) --------------------------------------
	authed := api.Group("", authRequired)
	authed.GET("/me", meHandler.Get)
	authed.PUT("/me", meHandler.Update)
	// Kontak WhatsApp mencatat pengguna yang menghubungi mitra; karenanya
	// endpoint ini mensyaratkan login (pengunjung diarahkan masuk lebih dulu).
	authed.GET("/properties/:slug/contact", catalogHandler.Contact)
	authed.POST("/partners/register", partnerHandler.Register)
	authed.POST("/subscriptions/checkout",
		middleware.RateLimit(cfg.RateLimit, cfg.RateLimit.CheckoutRequests),
		checkoutHandler.Checkout,
	)
	authed.GET("/subscriptions/orders/:orderCode/status", checkoutHandler.Status)
	authed.POST("/subscriptions/orders/:orderCode/charge",
		middleware.RateLimit(cfg.RateLimit, cfg.RateLimit.CheckoutRequests),
		checkoutHandler.Charge,
	)
	authed.GET("/subscriptions/orders/:orderCode/payment", checkoutHandler.Payment)
	authed.GET("/invoices/:orderCode", invoiceHandler.Get)
	authed.GET("/invoices/:orderCode/pdf", invoiceHandler.PDF)

	// --- Customer ----------------------------------------------------------
	customer := api.Group("", authRequired, middleware.RequireRole(models.RoleCustomer))
	customer.GET("/me/favorites", favoriteHandler.List)
	customer.POST("/me/favorites/:id", favoriteHandler.Add)
	customer.DELETE("/me/favorites/:id", favoriteHandler.Remove)

	// --- Mitra: profil dan perjanjian (tanpa syarat paket) -----------------
	mitraProfile := api.Group("/vendor", authRequired, middleware.RequireRole(models.RoleMitra))
	mitraProfile.GET("/profile", partnerHandler.Profile)
	mitraProfile.PUT("/profile", partnerHandler.UpdateProfile)
	mitraProfile.GET("/agreement", partnerHandler.Agreement)
	// Riwayat transaksi langganan mitra; tidak mensyaratkan paket aktif agar
	// mitra yang baru checkout tetap dapat melihat status pembayarannya.
	mitraProfile.GET("/orders", checkoutHandler.Orders)

	// --- Mitra: dashboard dan properti -------------------------------------
	vendor := api.Group("/vendor", authRequired, middleware.RequireRole(models.RoleMitra), activePartner)
	vendor.GET("/dashboard", vendorDashboardHandler.Dashboard)
	vendor.GET("/subscription", vendorSubscriptionHandler.Current)

	vendor.GET("/properties", vendorPropertyHandler.List)
	vendor.GET("/properties/:id", vendorPropertyHandler.Detail)
	vendor.POST("/properties", activePackage, vendorPropertyHandler.Create)
	vendor.PUT("/properties/:id", activePackage, vendorPropertyHandler.Update)
	vendor.DELETE("/properties/:id", activePackage, vendorPropertyHandler.Delete)
	vendor.PATCH("/properties/:id/status", activePackage, vendorPropertyHandler.SetStatus)
	vendor.PATCH("/properties/:id/availability", activePackage, vendorPropertyHandler.SetAvailability)
	// Boost listing dibatasi kuota paket aktif.
	vendor.PATCH("/properties/:id/boost", activePackage, vendorPropertyHandler.SetBoost)
	// Catatan lead (klik WhatsApp / isi form) untuk tabel properti mitra.
	vendor.GET("/properties/:id/leads", vendorPropertyHandler.Leads)
	vendor.DELETE("/properties/:id/photos/:photoId", activePackage, vendorMediaHandler.Delete)
	vendor.PATCH("/properties/:id/photos/order", activePackage, vendorMediaHandler.Order)
	vendor.PATCH("/properties/:id/photos/:photoId/primary", activePackage, vendorMediaHandler.SetPrimary)

	uploads.POST("/vendor/properties/:id/photos",
		authRequired,
		middleware.RequireRole(models.RoleMitra),
		activePartner,
		activePackage,
		middleware.UploadGuard(cfg.Storage),
		vendorMediaHandler.Upload,
	)

	// --- Admin dan Superadmin ----------------------------------------------
	admin := api.Group("/admin",
		authRequired,
		middleware.RequireRole(models.RoleAdmin, models.RoleSuperadmin),
		auditHook,
	)
	admin.GET("/dashboard", adminDashboardHandler.Dashboard)

	admin.GET("/partners", adminPartnerHandler.List)
	admin.GET("/partners/:id", adminPartnerHandler.Detail)
	admin.PATCH("/partners/:id/status", adminPartnerHandler.SetStatus)

	admin.GET("/properties", adminPropertyHandler.List)
	admin.POST("/properties", adminPropertyHandler.Create)
	admin.PUT("/properties/:id", adminPropertyHandler.Update)
	admin.PATCH("/properties/:id/status", adminPropertyHandler.SetStatus)
	admin.DELETE("/properties/:id/photos/:photoId", adminMediaHandler.Delete)

	admin.GET("/reports/subscriptions", adminReportHandler.Subscriptions)
	admin.GET("/reports/subscriptions/export", adminReportHandler.Export)
	admin.GET("/reports/subscriptions/:orderCode", adminReportHandler.Detail)

	// Admin hanya boleh membaca daftar paket; perubahan paket khusus superadmin.
	admin.GET("/packages", adminPlanHandler.List)

	// Master data kategori dapat dibaca admin, diubah khusus superadmin.
	admin.GET("/categories", adminCategoryHandler.List)

	// Konten landing page: admin dan superadmin boleh mengubah lalu publikasi.
	admin.GET("/landing", landingHandler.Get)
	admin.PUT("/landing", landingHandler.Update)
	admin.POST("/landing/publish", landingHandler.Publish)

	uploads.POST("/admin/properties/:id/photos",
		authRequired,
		middleware.RequireRole(models.RoleAdmin, models.RoleSuperadmin),
		auditHook,
		middleware.UploadGuard(cfg.Storage),
		adminMediaHandler.Upload,
	)

	// Unggah gambar landing (hero/og) dengan konversi otomatis ke WebP.
	uploads.POST("/admin/landing/images/:kind",
		authRequired,
		middleware.RequireRole(models.RoleAdmin, models.RoleSuperadmin),
		auditHook,
		middleware.UploadGuard(cfg.Storage),
		landingHandler.UploadImage,
	)

	// Unggah gambar pengaturan (logo situs / gambar area masuk) dengan konversi
	// otomatis ke WebP; URL hasil disimpan ke pengaturan terkait.
	uploads.POST("/admin/settings/images/:kind",
		authRequired,
		middleware.RequireRole(models.RoleSuperadmin),
		auditHook,
		middleware.UploadGuard(cfg.Storage),
		adminSettingHandler.UploadImage,
	)

	// --- Superadmin ---------------------------------------------------------
	super := api.Group("/admin", authRequired, middleware.RequireRole(models.RoleSuperadmin), auditHook)
	super.GET("/users", adminUserHandler.List)
	super.POST("/users", adminUserHandler.Create)
	super.GET("/users/:id", adminUserHandler.Detail)
	super.PUT("/users/:id", adminUserHandler.Update)
	super.PATCH("/users/:id/status", adminUserHandler.SetStatus)
	super.POST("/users/:id/reset-password", adminUserHandler.ResetPassword)

	super.POST("/packages", adminPlanHandler.Create)
	super.PUT("/packages/:id", adminPlanHandler.Update)
	super.PATCH("/packages/:id/status", adminPlanHandler.SetStatus)
	super.PATCH("/packages/order", adminPlanHandler.Reorder)

	// Master data kategori properti.
	super.POST("/categories", adminCategoryHandler.Create)
	super.PUT("/categories/:id", adminCategoryHandler.Update)
	super.PATCH("/categories/:id/status", adminCategoryHandler.SetStatus)
	super.DELETE("/categories/:id", adminCategoryHandler.Delete)

	super.GET("/agreements", adminAgreementHandler.List)
	super.POST("/agreements", adminAgreementHandler.Create)
	super.PATCH("/agreements/:id/activate", adminAgreementHandler.Activate)

	super.GET("/settings", adminSettingHandler.Get)
	super.PUT("/settings", adminSettingHandler.Update)

	// Audit log dapat dibaca superadmin dan owner (owner hanya metode GET).
	auditLogs := api.Group("/admin",
		authRequired,
		middleware.RequireRole(models.RoleSuperadmin, models.RoleOwner),
		auditHook,
	)
	auditLogs.GET("/audit-logs", adminAuditHandler.List)
	auditLogs.GET("/audit-logs/export", adminAuditHandler.Export)

	// --- Owner --------------------------------------------------------------
	owner := api.Group("/owner", authRequired, middleware.RequireRole(models.RoleOwner))
	owner.GET("/dashboard", ownerHandler.Dashboard)
	owner.GET("/reports/subscriptions", ownerHandler.Subscriptions)
	owner.GET("/reports/partners", ownerHandler.Partners)
	// Owner hanya dapat melihat konten landing page (tanpa hak ubah).
	owner.GET("/landing", landingHandler.Get)

	return engine
}
