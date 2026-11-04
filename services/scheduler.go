package services

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"

	"enmproperty/backend/platform/logger"
)

// SchedulerService menjalankan job terjadwal (BE-16): kedaluwarsa langganan
// harian dan pembersihan order yang belum dibayar. Setiap job idempoten
// sehingga aman dijalankan ulang maupun saat proses dijalankan berkali-kali.
type SchedulerService struct {
	deps          Deps
	subscriptions *SubscriptionService
	payments      *PaymentService
	logger        zerolog.Logger
	cron          *cron.Cron
}

// Start mendaftarkan seluruh job sesuai jadwal pada konfigurasi lalu
// menjalankan scheduler di latar belakang.
func (s *SchedulerService) Start() error {
	if !s.deps.Config.Scheduler.Enabled {
		s.logger.Info().Msg("scheduler dinonaktifkan melalui konfigurasi")
		return nil
	}
	if s.cron != nil {
		return nil
	}

	location, err := s.deps.Config.Location()
	if err != nil {
		return fmt.Errorf("gagal memuat zona waktu scheduler: %w", err)
	}

	runner := cron.New(cron.WithLocation(location))

	if _, err := runner.AddFunc(s.deps.Config.Scheduler.ExpireSubscriptionsCron, func() {
		s.ExpireSubscriptions(timeNow())
	}); err != nil {
		return fmt.Errorf("jadwal SCHEDULER_EXPIRE_SUBSCRIPTIONS_CRON tidak valid: %w", err)
	}
	if _, err := runner.AddFunc(s.deps.Config.Scheduler.ExpireOrdersCron, func() {
		s.ExpireOrders(timeNow())
	}); err != nil {
		return fmt.Errorf("jadwal SCHEDULER_EXPIRE_ORDERS_CRON tidak valid: %w", err)
	}

	runner.Start()
	s.cron = runner

	s.logger.Info().
		Str("expire_subscriptions", s.deps.Config.Scheduler.ExpireSubscriptionsCron).
		Str("expire_orders", s.deps.Config.Scheduler.ExpireOrdersCron).
		Msg("scheduler dijalankan")
	return nil
}

// Stop menghentikan scheduler dan menunggu job yang sedang berjalan selesai
// atau konteks dibatalkan.
func (s *SchedulerService) Stop(ctx context.Context) {
	if s.cron == nil {
		return
	}
	stopped := s.cron.Stop()
	s.cron = nil

	select {
	case <-stopped.Done():
	case <-ctx.Done():
	}
	s.logger.Info().Msg("scheduler dihentikan")
}

// ExpireSubscriptions menandai langganan yang lewat masa aktif menjadi
// kedaluwarsa, menonaktifkan mitra, dan menonaktifkan propertinya.
func (s *SchedulerService) ExpireSubscriptions(now time.Time) int {
	count, err := s.subscriptions.ExpireDue(now)
	if err != nil {
		s.logger.Error().Err(err).Msg("job kedaluwarsa langganan gagal")
		return 0
	}
	s.logger.Info().Int("expired", count).Msg("job kedaluwarsa langganan selesai")
	return count
}

// ExpireOrders menandai order yang melewati batas waktu pembayaran menjadi
// kedaluwarsa sehingga slot kuota mitra dapat dipakai kembali.
func (s *SchedulerService) ExpireOrders(now time.Time) int {
	count, err := s.payments.ExpireStaleOrders(now)
	if err != nil {
		s.logger.Error().Err(err).Msg("job kedaluwarsa order gagal")
		return 0
	}
	s.logger.Info().Int("expired", count).Msg("job kedaluwarsa order selesai")
	return count
}

// newSchedulerLogger menyiapkan logger khusus scheduler dari konfigurasi log.
func newSchedulerLogger(deps Deps) zerolog.Logger {
	return logger.New(deps.Config.Log).With().Str("component", "scheduler").Logger()
}
