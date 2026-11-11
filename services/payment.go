package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/pkg/money"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/platform/midtrans"
)

// PaymentService mengelola checkout paket, webhook Midtrans, status order, dan
// invoice langganan (BE-06).
type PaymentService struct {
	deps          Deps
	partners      *PartnerService
	agreements    *AgreementService
	subscriptions *SubscriptionService
	settings      *SettingService
}

// Checkout membuat order langganan untuk mitra, melengkapi perjanjian bila
// diwajibkan, dan meminta token Snap dari Midtrans (atau mode simulasi).
func (s *PaymentService) Checkout(ctx context.Context, user *models.User, req dto.CheckoutRequest, meta RequestMeta) (dto.CheckoutResponse, error) {
	if user == nil {
		return dto.CheckoutResponse{}, httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}

	plan, err := s.deps.Repos.Plans.GetBySlug(strings.TrimSpace(req.PackageSlug))
	if err != nil {
		if database.IsNotFound(err) {
			return dto.CheckoutResponse{}, httpx.NotFound("Paket tidak ditemukan")
		}
		return dto.CheckoutResponse{}, httpx.Internal(err)
	}
	if !plan.IsActive {
		return dto.CheckoutResponse{}, httpx.Unprocessable("Paket sedang tidak tersedia")
	}

	partner, err := s.resolvePartner(ctx, user, req.Partner)
	if err != nil {
		return dto.CheckoutResponse{}, err
	}

	now := timeNow()
	kind := s.resolveKind(partner.ID, plan, now)

	// Paket gratis langsung aktif tanpa pembayaran sehingga mitra dapat segera
	// memasang listing. Bila mitra sudah menyetujui perjanjian pada dialog
	// konfirmasi, persetujuannya tetap dicatat (best-effort) agar halaman
	// "Perjanjian Kemitraan" menampilkan status disetujui; kegagalan
	// pencatatan tidak menahan aktivasi paket Gratis.
	if plan.IsFree {
		if req.AgreeAgreement && s.settings.AgreementRequired() {
			_, _ = s.agreements.Accept(user, partner, req.AgreementID, meta)
		}

		order := &models.Order{
			ID:               newID(),
			UserID:           user.ID,
			PartnerID:        partner.ID,
			PartnerName:      partner.BusinessName,
			PlanID:           plan.ID,
			PlanName:         plan.Name,
			PlanPrice:        plan.Price,
			PlanDurationDays: plan.DurationDays,
			PlanListingQuota: plan.ListingQuota,
			Amount:           plan.Price,
			Kind:             kind,
			Status:           models.OrderStatusPaid,
			Provider:         models.PaymentProviderFree,
			PaidAt:           &now,
			ExpiresAt:        now.Add(time.Duration(s.orderExpiryMinutes()) * time.Minute),
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := s.createOrder(order); err != nil {
			return dto.CheckoutResponse{}, err
		}
		s.fulfill(ctx, order)
		return dto.CheckoutResponse{
			Order:      dto.NewOrderView(order),
			FreeActive: true,
		}, nil
	}

	if s.settings.AgreementRequired() {
		if !req.AgreeAgreement {
			return dto.CheckoutResponse{}, httpx.Unprocessable("Anda harus menyetujui perjanjian kemitraan terlebih dahulu")
		}
		if _, err := s.agreements.Accept(user, partner, req.AgreementID, meta); err != nil {
			return dto.CheckoutResponse{}, err
		}
	}

	order := &models.Order{
		ID:               newID(),
		UserID:           user.ID,
		PartnerID:        partner.ID,
		PartnerName:      partner.BusinessName,
		PlanID:           plan.ID,
		PlanName:         plan.Name,
		PlanPrice:        plan.Price,
		PlanDurationDays: plan.DurationDays,
		PlanListingQuota: plan.ListingQuota,
		Amount:           plan.Price,
		Kind:             kind,
		Status:           models.OrderStatusPending,
		Provider:         models.PaymentProviderMidtrans,
		ExpiresAt:        now.Add(time.Duration(s.orderExpiryMinutes()) * time.Minute),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.createOrder(order); err != nil {
		return dto.CheckoutResponse{}, err
	}

	snap, err := s.deps.Midtrans.CreateSnapToken(ctx, midtrans.Request{
		OrderID:       order.Code,
		Amount:        order.Amount,
		CustomerName:  user.DisplayName(),
		CustomerEmail: user.Email,
		CustomerPhone: partner.Phone,
		Items: []midtrans.Item{{
			ID:       plan.ID,
			Name:     plan.Name,
			Price:    order.Amount,
			Quantity: 1,
		}},
		ExpiryMinutes: s.orderExpiryMinutes(),
	})
	if err != nil {
		order.Status = models.OrderStatusFailed
		order.UpdatedAt = timeNow()
		_ = s.deps.Repos.Orders.Update(order)
		return dto.CheckoutResponse{}, httpx.Internal(err)
	}

	order.SnapToken = snap.Token
	order.SnapRedirectURL = snap.RedirectURL
	if snap.Simulated {
		order.Provider = models.PaymentProviderSimulated
	}
	order.UpdatedAt = timeNow()
	if err := s.deps.Repos.Orders.Update(order); err != nil {
		return dto.CheckoutResponse{}, httpx.Internal(err)
	}

	return dto.CheckoutResponse{
		Order:           dto.NewOrderView(order),
		SnapToken:       snap.Token,
		SnapRedirectURL: snap.RedirectURL,
		ClientKey:       s.deps.Config.Midtrans.ClientKey,
		Simulated:       snap.Simulated,
	}, nil
}

// HandleNotification memproses webhook Midtrans: verifikasi signature,
// idempotensi, dan pemetaan status transaksi ke status order.
func (s *PaymentService) HandleNotification(ctx context.Context, n dto.MidtransNotification) error {
	order, err := s.deps.Repos.Orders.GetByCode(strings.TrimSpace(n.OrderID))
	if err != nil {
		// Order tidak dikenal: abaikan agar Midtrans tidak mengulang tanpa henti.
		if database.IsNotFound(err) {
			return nil
		}
		return httpx.Internal(err)
	}
	if order.IsSettled() {
		return nil
	}
	if !s.deps.Midtrans.VerifySignature(order.Code, n.StatusCode, n.GrossAmount, n.SignatureKey) {
		return httpx.Unauthorized("Signature webhook tidak valid")
	}

	status := mapMidtransStatus(n.TransactionStatus, n.FraudStatus)
	if status == "" || status == order.Status {
		return nil
	}
	return s.applyStatus(ctx, order, status, n)
}

// GetStatus mengembalikan status order milik pengguna beserta langganannya.
func (s *PaymentService) GetStatus(user *models.User, orderCode string) (dto.OrderStatusView, error) {
	order, err := s.orderByCode(orderCode)
	if err != nil {
		return dto.OrderStatusView{}, err
	}
	if !canViewOrder(user, order) {
		return dto.OrderStatusView{}, httpx.Forbidden("Anda tidak berhak melihat order ini")
	}
	return s.statusView(order), nil
}

// Charge membuat transaksi QRIS/VA melalui Midtrans Core API lalu menyimpan
// instruksi pembayaran pada order agar dapat dirender di halaman pembayaran.
func (s *PaymentService) Charge(ctx context.Context, user *models.User, orderCode, method string) (dto.PaymentView, error) {
	order, err := s.orderByCode(orderCode)
	if err != nil {
		return dto.PaymentView{}, err
	}
	if !canViewOrder(user, order) {
		return dto.PaymentView{}, httpx.Forbidden("Anda tidak berhak membayar order ini")
	}
	if order.IsPaid() {
		return s.paymentView(order), nil
	}
	if order.Status != models.OrderStatusPending {
		return dto.PaymentView{}, httpx.Unprocessable("Order tidak dapat dibayar pada status ini")
	}
	if order.ExpiresAt.Before(timeNow()) {
		return dto.PaymentView{}, httpx.Unprocessable("Batas waktu pembayaran telah lewat")
	}

	method = normalizePaymentMethod(method)
	request := midtrans.Request{
		OrderID:       order.Code,
		Amount:        order.Amount,
		CustomerName:  user.DisplayName(),
		CustomerEmail: user.Email,
		ExpiryMinutes: s.orderExpiryMinutes(),
	}
	if order.PartnerID != "" {
		if partner, err := s.deps.Repos.Partners.GetByID(order.PartnerID); err == nil && partner != nil {
			request.CustomerPhone = partner.Phone
		}
	}

	var result *midtrans.ChargeResult
	if method == "qris" {
		result, err = s.deps.Midtrans.ChargeQRIS(ctx, request)
	} else {
		result, err = s.deps.Midtrans.ChargeVA(ctx, method, request)
	}
	if err != nil {
		return dto.PaymentView{}, chargeError(err)
	}

	s.applyChargeResult(order, method, result)
	if err := s.deps.Repos.Orders.Update(order); err != nil {
		return dto.PaymentView{}, httpx.Internal(err)
	}
	return s.paymentView(order), nil
}

// chargeError menerjemahkan kegagalan charge Midtrans menjadi respons yang
// dapat ditindaklanjuti. Pesan asli Midtrans (mis. "Payment channel is not
// activated") diteruskan dengan status 422 agar penyebabnya terlihat; bila
// Midtrans justru melaporkan gangguan sisi layanan (5xx), pesan tetap
// disamarkan melalui httpx.Internal.
func chargeError(err error) *httpx.AppError {
	if status, message, ok := midtrans.ErrorDetail(err); ok {
		if message == "" {
			message = "Transaksi pembayaran ditolak oleh Midtrans"
		}
		if status >= 500 {
			return httpx.Internal(err)
		}
		return httpx.Unprocessable(message)
	}
	return httpx.Internal(err)
}

// GetPayment mengembalikan instruksi pembayaran yang sudah dibuat untuk order.
func (s *PaymentService) GetPayment(user *models.User, orderCode string) (dto.PaymentView, error) {
	order, err := s.orderByCode(orderCode)
	if err != nil {
		return dto.PaymentView{}, err
	}
	if !canViewOrder(user, order) {
		return dto.PaymentView{}, httpx.Forbidden("Anda tidak berhak melihat pembayaran ini")
	}
	return s.paymentView(order), nil
}

// ListOrders mengembalikan riwayat transaksi milik pengguna (mitra) berpaginasi.
func (s *PaymentService) ListOrders(user *models.User, q dto.PaginationQuery, extra map[string]string) (dto.ListResult[dto.OrderView], error) {
	if user == nil {
		return dto.ListResult[dto.OrderView]{}, httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}
	filters := map[string]string{"user_id": user.ID}
	for key, value := range extra {
		if strings.TrimSpace(value) != "" {
			filters[key] = value
		}
	}

	orders, total, err := s.deps.Repos.Orders.List(q.ToListQuery(filters))
	if err != nil {
		return dto.ListResult[dto.OrderView]{}, httpx.Internal(err)
	}
	items := make([]dto.OrderView, 0, len(orders))
	for _, order := range orders {
		items = append(items, dto.NewOrderView(order))
	}
	return dto.NewListResult(items, total), nil
}

// paymentView melengkapi instruksi pembayaran dengan flag retry berbasis waktu.
func (s *PaymentService) paymentView(order *models.Order) dto.PaymentView {
	view := dto.NewPaymentView(order)
	view.CanRetry = order.Status == models.OrderStatusPending && !order.ExpiresAt.Before(timeNow())
	return view
}

// applyChargeResult menyimpan hasil charge Core API pada entitas order.
func (s *PaymentService) applyChargeResult(order *models.Order, method string, result *midtrans.ChargeResult) {
	if method == "qris" {
		order.PaymentMethod = "qris"
		order.PaymentChannel = "qris"
		order.PaymentQRURL = result.QRURL
		order.PaymentQRString = result.QRString
	} else {
		order.PaymentMethod = "va"
		order.PaymentChannel = method
		order.PaymentNumber = result.VANumber
	}
	if result.TransactionID != "" {
		order.MidtransTransactionID = result.TransactionID
	}
	if result.PaymentType != "" {
		order.MidtransPaymentType = result.PaymentType
	}

	expires := order.ExpiresAt
	if parsed, err := time.Parse("2006-01-02 15:04:05", result.ExpiryTime); err == nil {
		expires = parsed
	}
	order.PaymentExpiresAt = &expires
	order.UpdatedAt = timeNow()
}

// normalizePaymentMethod menormalkan nama metode pembayaran dari klien.
func normalizePaymentMethod(method string) string {
	return strings.ToLower(strings.TrimSpace(method))
}

// SyncStatusFromMidtrans menarik status terkini dari Midtrans lalu menyimpannya.
func (s *PaymentService) SyncStatusFromMidtrans(ctx context.Context, orderCode string) (dto.OrderStatusView, error) {
	order, err := s.orderByCode(orderCode)
	if err != nil {
		return dto.OrderStatusView{}, err
	}
	if order.IsSettled() {
		return s.statusView(order), nil
	}

	result, err := s.deps.Midtrans.TransactionStatus(ctx, order.Code)
	if err != nil {
		return dto.OrderStatusView{}, httpx.Internal(err)
	}
	status := mapMidtransStatus(result.TransactionStatus, result.FraudStatus)
	if status != "" && status != order.Status {
		notification := dto.MidtransNotification{
			OrderID:           order.Code,
			StatusCode:        result.StatusCode,
			GrossAmount:       result.GrossAmount,
			TransactionID:     result.TransactionID,
			TransactionStatus: result.TransactionStatus,
			FraudStatus:       result.FraudStatus,
			PaymentType:       result.PaymentType,
		}
		if err := s.applyStatus(ctx, order, status, notification); err != nil {
			return dto.OrderStatusView{}, err
		}
		order, _ = s.deps.Repos.Orders.GetByID(order.ID)
	}
	return s.statusView(order), nil
}

// GetInvoice membentuk data invoice untuk satu order langganan.
func (s *PaymentService) GetInvoice(user *models.User, orderCode string) (dto.InvoiceView, error) {
	order, err := s.orderByCode(orderCode)
	if err != nil {
		return dto.InvoiceView{}, err
	}
	if !canViewOrder(user, order) {
		return dto.InvoiceView{}, httpx.Forbidden("Anda tidak berhak melihat invoice ini")
	}

	item := dto.InvoiceItem{
		Name:        order.PlanName,
		Description: fmt.Sprintf("Langganan paket %d hari", order.PlanDurationDays),
		Quantity:    1,
		UnitPrice:   order.Amount,
		Total:       order.Amount,
		TotalLabel:  money.Rupiah(order.Amount),
	}

	view := dto.InvoiceView{
		InvoiceNumber: order.Code,
		Order:         dto.NewOrderView(order),
		Items:         []dto.InvoiceItem{item},
		Subtotal:      order.Amount,
		SubtotalLabel: money.Rupiah(order.Amount),
		Total:         order.Amount,
		TotalLabel:    money.Rupiah(order.Amount),
		IssuedAt:      order.CreatedAt,
		PaidAt:        order.PaidAt,
		IsPaid:        order.IsPaid(),
	}
	if partner, err := s.deps.Repos.Partners.GetByID(order.PartnerID); err == nil && partner != nil {
		view.Partner = dto.PartnerDetail{
			ID:           partner.ID,
			BusinessName: partner.BusinessName,
			PICName:      partner.PICName,
			Phone:        partner.Phone,
			Email:        partner.Email,
			Address:      partner.Address,
			City:         partner.City,
			Province:     partner.Province,
		}
	}
	return view, nil
}

// ExpireStaleOrders menandai order yang belum dibayar dan sudah lewat batas
// waktu sebagai kedaluwarsa. Idempoten.
func (s *PaymentService) ExpireStaleOrders(now time.Time) (int, error) {
	orders, _, err := s.deps.Repos.Orders.List(database.ListQuery{
		PerPage: 0,
		Filters: map[string]string{"status": string(models.OrderStatusPending)},
	})
	if err != nil {
		return 0, httpx.Internal(err)
	}

	expired := 0
	for _, order := range orders {
		if now.Before(order.ExpiresAt) {
			continue
		}
		order.Status = models.OrderStatusExpired
		order.UpdatedAt = now
		if err := s.deps.Repos.Orders.Update(order); err != nil {
			return expired, httpx.Internal(err)
		}
		expired++
	}
	return expired, nil
}

// resolvePartner mengambil profil mitra pengguna.
//
// Bila profil mitra belum ada, profil awal dibuat dari data akun sehingga
// checkout dapat berjalan tanpa mengisi data badan usaha lebih dulu; data bisnis
// dilengkapi mitra setelah pembayaran berhasil. Data pendaftaran (`req`) tetap
// dipakai bila klien mengirimkannya lebih awal.
func (s *PaymentService) resolvePartner(ctx context.Context, user *models.User, req *dto.RegisterPartnerRequest) (*models.Partner, error) {
	if req != nil {
		return s.partners.Ensure(ctx, user, *req)
	}
	partner, err := s.deps.Repos.Partners.GetByUserID(user.ID)
	if err == nil {
		return partner, nil
	}
	if !database.IsNotFound(err) {
		return nil, httpx.Internal(err)
	}
	return s.partners.EnsureFromUser(ctx, user)
}

// resolveKind menentukan jenis order: baru, perpanjangan, atau upgrade.
func (s *PaymentService) resolveKind(partnerID string, plan *models.Plan, now time.Time) models.OrderKind {
	existing := s.subscriptions.Active(partnerID)
	if existing == nil || !existing.IsActiveAt(now) {
		return models.OrderKindNew
	}
	if existing.PlanID == plan.ID {
		return models.OrderKindRenewal
	}
	currentPrice := int64(0)
	if current, err := s.deps.Repos.Plans.GetByID(existing.PlanID); err == nil && current != nil {
		currentPrice = current.Price
	}
	if plan.Price > currentPrice {
		return models.OrderKindUpgrade
	}
	return models.OrderKindRenewal
}

// createOrder menyimpan order sambil memastikan kode order unik.
func (s *PaymentService) createOrder(order *models.Order) error {
	for attempt := 0; attempt < 5; attempt++ {
		order.Code = s.nextOrderCode(order.CreatedAt)
		err := s.deps.Repos.Orders.Create(order)
		if err == nil {
			return nil
		}
		if database.IsDuplicateCode(err) {
			continue
		}
		return httpx.Internal(err)
	}
	return httpx.Conflict("Gagal membuat kode order yang unik, silakan coba lagi")
}

// nextOrderCode membentuk kode order berurutan per hari.
func (s *PaymentService) nextOrderCode(now time.Time) string {
	prefix := "INV-" + now.Format("20060102")
	sequence := s.deps.Repos.Orders.CountByCodePrefix(prefix) + 1
	return fmt.Sprintf("%s-%04d", prefix, sequence)
}

// orderExpiryMinutes mengembalikan masa berlaku order dalam menit.
func (s *PaymentService) orderExpiryMinutes() int {
	minutes := s.deps.Config.Midtrans.OrderExpiryMinutes
	if minutes <= 0 {
		return 60
	}
	return minutes
}

// orderByCode mengambil order berdasarkan kode.
func (s *PaymentService) orderByCode(orderCode string) (*models.Order, error) {
	order, err := s.deps.Repos.Orders.GetByCode(strings.TrimSpace(orderCode))
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Order tidak ditemukan")
		}
		return nil, httpx.Internal(err)
	}
	return order, nil
}

// statusView melengkapi status order dengan langganan aktif mitra.
func (s *PaymentService) statusView(order *models.Order) dto.OrderStatusView {
	view := dto.OrderStatusView{
		Order:    dto.NewOrderView(order),
		CanRetry: order.Status == models.OrderStatusPending && !order.ExpiresAt.Before(timeNow()),
	}
	if partner, err := s.deps.Repos.Partners.GetByID(order.PartnerID); err == nil && partner != nil {
		if subscription, err := s.subscriptions.GetCurrent(partner); err == nil {
			view.Subscription = subscription
		}
	}
	return view
}

// applyStatus menyimpan perubahan status order dan menjalankan pemenuhan bila lunas.
func (s *PaymentService) applyStatus(ctx context.Context, order *models.Order, status models.OrderStatus, n dto.MidtransNotification) error {
	now := timeNow()
	order.Status = status
	order.UpdatedAt = now
	if n.TransactionID != "" {
		order.MidtransTransactionID = n.TransactionID
	}
	if n.PaymentType != "" {
		order.MidtransPaymentType = n.PaymentType
	}
	if status == models.OrderStatusPaid && order.PaidAt == nil {
		order.PaidAt = &now
	}
	if err := s.deps.Repos.Orders.Update(order); err != nil {
		return httpx.Internal(err)
	}
	if status == models.OrderStatusPaid {
		s.fulfill(ctx, order)
	}
	return nil
}

// fulfill mengaktifkan langganan dan kemitraan setelah pembayaran lunas.
func (s *PaymentService) fulfill(_ context.Context, order *models.Order) {
	if _, err := s.subscriptions.ActivateFromOrder(order); err != nil {
		return
	}
	s.activatePartner(order.PartnerID)
}

// activatePartner mengaktifkan kemitraan yang masih menunggu atau kedaluwarsa.
func (s *PaymentService) activatePartner(partnerID string) {
	partner, err := s.deps.Repos.Partners.GetByID(partnerID)
	if err != nil || partner == nil {
		return
	}
	if partner.Status == models.PartnerStatusActive || partner.Status == models.PartnerStatusSuspended {
		return
	}

	now := timeNow()
	partner.Status = models.PartnerStatusActive
	partner.SuspendedReason = ""
	if partner.JoinedAt == nil {
		partner.JoinedAt = &now
	}
	partner.UpdatedAt = now
	if err := s.deps.Repos.Partners.Update(partner); err != nil {
		return
	}
	s.subscriptions.RestoreProperties(partnerID)
}

// mapMidtransStatus memetakan status transaksi Midtrans ke status order.
func mapMidtransStatus(transactionStatus, fraudStatus string) models.OrderStatus {
	switch strings.ToLower(strings.TrimSpace(transactionStatus)) {
	case "capture", "settlement":
		if strings.EqualFold(fraudStatus, "deny") {
			return models.OrderStatusFailed
		}
		return models.OrderStatusPaid
	case "deny":
		return models.OrderStatusFailed
	case "cancel":
		return models.OrderStatusCancelled
	case "expire":
		return models.OrderStatusExpired
	case "pending":
		return models.OrderStatusPending
	default:
		return ""
	}
}

// canViewOrder memeriksa apakah pengguna berhak melihat order.
func canViewOrder(user *models.User, order *models.Order) bool {
	if user == nil || order == nil {
		return false
	}
	return order.UserID == user.ID || user.Role.IsInternal()
}
