// Command enmproperty-api menjalankan HTTP server backend ENMProperty.
//
// Alur startup: memuat konfigurasi, menyiapkan logger, lapisan data, integrasi
// platform (Firebase, Midtrans, penyimpanan media), merakit service dan router,
// lalu menjalankan server dengan graceful shutdown beserta job terjadwal.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"enmproperty/backend/config"
	"enmproperty/backend/database"
	"enmproperty/backend/migrations"
	"enmproperty/backend/platform/firebase"
	"enmproperty/backend/platform/logger"
	"enmproperty/backend/platform/midtrans"
	"enmproperty/backend/platform/storage"
	"enmproperty/backend/router"
	"enmproperty/backend/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gagal memuat konfigurasi: %v", err)
	}

	appLogger := logger.New(cfg.Log)
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()

	// Lapisan data dipilih berdasarkan DB_DRIVER. Repository berada di balik
	// interface sehingga service maupun handler tidak terpengaruh.
	var (
		pool  *pgxpool.Pool
		repos database.Repositories
	)

	if cfg.UseMemoryStore() {
		// Mode in-memory: data contoh dimuat agar pengembangan tanpa PostgreSQL
		// tetap dapat berjalan.
		store := database.NewStore()
		database.Seed(store)
		repos = database.NewMemoryRepositories(store)
	} else {
		// Mode PostgreSQL: siapkan pool pgx dan (bila diminta) terapkan migrasi.
		pool, err = database.NewPostgresPool(ctx, cfg.Database)
		if err != nil {
			appLogger.Fatal().Err(err).Msg("gagal menyiapkan pool PostgreSQL")
		}
		defer pool.Close()

		if cfg.Database.AutoMigrate {
			if err := migrations.Run(ctx, pool); err != nil {
				appLogger.Fatal().Err(err).Msg("gagal menjalankan migrasi basis data")
			}
		}
		repos = database.NewPostgresRepositories(pool)
	}

	firebaseClient, err := firebase.New(ctx, cfg.Firebase, cfg.Auth)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("gagal menyiapkan klien Firebase")
	}
	midtransClient := midtrans.New(cfg.Midtrans)

	mediaStorage, err := storage.New(cfg.Storage, cfg.Firebase)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("gagal menyiapkan penyimpanan media")
	}

	svc := services.NewServices(services.Deps{
		Repos:    repos,
		Config:   cfg,
		Firebase: firebaseClient,
		Midtrans: midtransClient,
		Storage:  mediaStorage,
	})

	engine := router.New(router.Deps{
		Config:   cfg,
		Services: svc,
		Repos:    repos,
		Firebase: firebaseClient,
		Midtrans: midtransClient,
		Logger:   appLogger,
	})

	// Foto yang disimpan pada driver lokal disajikan langsung oleh server.
	if strings.EqualFold(cfg.Storage.Driver, "local") {
		engine.Static("/uploads", cfg.Storage.LocalDir)
	}

	if err := svc.Scheduler.Start(); err != nil {
		appLogger.Error().Err(err).Msg("scheduler gagal dijalankan")
	}

	server := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      engine,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		appLogger.Info().
			Str("port", cfg.HTTP.Port).
			Str("env", cfg.App.Env).
			Bool("memory_store", cfg.UseMemoryStore()).
			Msg("server enmproperty berjalan")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal().Err(err).Msg("server berhenti karena kesalahan")
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error().Err(err).Msg("gagal menghentikan server dengan rapi")
	}
	svc.Scheduler.Stop(shutdownCtx)
	appLogger.Info().Msg("server dihentikan")
}
