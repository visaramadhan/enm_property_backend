package services

import (
	"bytes"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/xuri/excelize/v2"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/pkg/money"
	"enmproperty/backend/platform/httpx"
)

// ReportService menyusun ringkasan dashboard dan laporan langganan untuk admin,
// owner, dan mitra, termasuk ekspor XLSX dan PDF (BE-15).
type ReportService struct {
	deps          Deps
	subscriptions *SubscriptionService
	properties    *PropertyService
}

// AdminDashboard menyusun ringkasan operasional untuk halaman admin.
func (s *ReportService) AdminDashboard() (dto.AdminDashboard, error) {
	partners, err := s.partnerStats()
	if err != nil {
		return dto.AdminDashboard{}, err
	}
	properties, err := s.propertyStats()
	if err != nil {
		return dto.AdminDashboard{}, err
	}
	subscriptions, err := s.subscriptionStats()
	if err != nil {
		return dto.AdminDashboard{}, err
	}
	orders, err := s.orderStats()
	if err != nil {
		return dto.AdminDashboard{}, err
	}

	return dto.AdminDashboard{
		Partners:         partners,
		Properties:       properties,
		Subscriptions:    subscriptions,
		Orders:           orders,
		PropertyViews:    len(s.deps.Repos.Analytics.Views()),
		PropertyContacts: len(s.deps.Repos.Analytics.Contacts()),
		GeneratedAt:      timeNow(),
	}, nil
}

// VendorDashboard menyusun ringkasan untuk halaman dashboard mitra.
func (s *ReportService) VendorDashboard(partner *models.Partner) (dto.VendorDashboard, error) {
	if partner == nil {
		return dto.VendorDashboard{}, httpx.BadRequest("Profil kemitraan tidak tersedia")
	}

	properties, total, err := s.deps.Repos.Properties.List(database.ListQuery{
		PerPage: 0,
		Filters: map[string]string{"partner_id": partner.ID},
	})
	if err != nil {
		return dto.VendorDashboard{}, httpx.Internal(err)
	}

	views, err := s.viewCount(partner.ID)
	if err != nil {
		return dto.VendorDashboard{}, err
	}
	contacts, err := s.contactCount(partner.ID)
	if err != nil {
		return dto.VendorDashboard{}, err
	}

	favorites := 0
	for _, property := range properties {
		favorites += s.deps.Repos.Favorites.CountByProperty(property.ID)
	}

	dashboard := dto.VendorDashboard{
		PartnerID:        partner.ID,
		PartnerName:      partner.BusinessName,
		PartnerStatus:    string(partner.Status),
		Properties:       countPropertyStatuses(properties),
		PropertyViews:    views[partner.ID],
		PropertyContacts: contacts[partner.ID],
		Favorites:        favorites,
		RecentProperties: s.recentProperties(properties, 5),
		GeneratedAt:      timeNow(),
	}

	if subscription, err := s.subscriptions.GetCurrent(partner); err == nil && subscription != nil {
		dashboard.Subscription = subscription
		dashboard.ListingsQuota = subscription.ListingQuota
		dashboard.ListingsUsed = subscription.ListingsUsed
		dashboard.ListingsRemaining = subscription.ListingRemaining
	}
	_ = total
	return dashboard, nil
}

// OwnerDashboard menyusun ringkasan untuk halaman owner beserta mitra teratas.
func (s *ReportService) OwnerDashboard() (dto.OwnerDashboard, error) {
	partners, err := s.partnerStats()
	if err != nil {
		return dto.OwnerDashboard{}, err
	}
	properties, err := s.propertyStats()
	if err != nil {
		return dto.OwnerDashboard{}, err
	}
	subscriptions, err := s.subscriptionStats()
	if err != nil {
		return dto.OwnerDashboard{}, err
	}
	orders, err := s.orderStats()
	if err != nil {
		return dto.OwnerDashboard{}, err
	}
	rows, err := s.partnerReportRows(dto.ReportFilterQuery{})
	if err != nil {
		return dto.OwnerDashboard{}, err
	}

	limit := 5
	if len(rows) < limit {
		limit = len(rows)
	}
	return dto.OwnerDashboard{
		Partners:      partners,
		Properties:    properties,
		Subscriptions: subscriptions,
		Orders:        orders,
		TopPartners:   rows[:limit],
		GeneratedAt:   timeNow(),
	}, nil
}

// OwnerPartnerReport menyusun laporan mitra beserta aktivitas propertinya.
func (s *ReportService) OwnerPartnerReport(q dto.ReportFilterQuery) (dto.ListResult[dto.PartnerReportRow], error) {
	rows, err := s.partnerReportRows(q)
	if err != nil {
		return dto.ListResult[dto.PartnerReportRow]{}, err
	}
	total := len(rows)
	page, _ := database.Paginate(rows, database.ListQuery{Page: q.Page, PerPage: q.PerPage})
	return dto.NewListResult(page, total), nil
}

// SubscriptionReport menyusun laporan transaksi langganan beserta totalnya.
func (s *ReportService) SubscriptionReport(q dto.ReportFilterQuery) (dto.SubscriptionReport, error) {
	rows, err := s.subscriptionReportRows(q)
	if err != nil {
		return dto.SubscriptionReport{}, err
	}

	total := len(rows)
	var totalAmount int64
	for _, row := range rows {
		if row.Status == string(models.OrderStatusPaid) {
			totalAmount += row.Amount
		}
	}
	return dto.SubscriptionReport{
		Rows:             rows,
		Total:            total,
		TotalAmount:      totalAmount,
		TotalAmountLabel: money.Rupiah(totalAmount),
	}, nil
}

// SubscriptionDetail mengembalikan satu baris laporan langganan berdasarkan kode order.
func (s *ReportService) SubscriptionDetail(orderCode string) (dto.SubscriptionReportRow, error) {
	order, err := s.deps.Repos.Orders.GetByCode(strings.TrimSpace(orderCode))
	if err != nil {
		if database.IsNotFound(err) {
			return dto.SubscriptionReportRow{}, httpx.NotFound("Order tidak ditemukan")
		}
		return dto.SubscriptionReportRow{}, httpx.Internal(err)
	}
	return subscriptionRow(order), nil
}

// ExportXLSX membentuk berkas Excel dari laporan langganan.
func (s *ReportService) ExportXLSX(q dto.ReportFilterQuery) ([]byte, string, error) {
	rows, err := s.subscriptionReportRows(q)
	if err != nil {
		return nil, "", err
	}

	file := excelize.NewFile()
	defer file.Close()
	const sheet = "Langganan"
	index, err := file.NewSheet(sheet)
	if err != nil {
		return nil, "", httpx.Internal(err)
	}
	file.SetActiveSheet(index)
	_ = file.DeleteSheet("Sheet1")

	headers := []string{"Kode Order", "Mitra", "Paket", "Jenis", "Status", "Penyedia", "Jumlah", "Dibuat", "Dibayar", "Berakhir"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = file.SetCellValue(sheet, cell, header)
	}
	for r, row := range rows {
		values := []any{
			row.OrderCode, row.PartnerName, row.PlanName, row.Kind, row.StatusLabel,
			row.Provider, row.Amount, row.CreatedAt.Format("2006-01-02 15:04"),
			formatTimePtr(row.PaidAt), row.ExpiresAt.Format("2006-01-02"),
		}
		for c, value := range values {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = file.SetCellValue(sheet, cell, value)
		}
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, "", httpx.Internal(err)
	}
	return buffer.Bytes(), "laporan-langganan.xlsx", nil
}

// ExportPDF membentuk berkas PDF dari laporan langganan.
func (s *ReportService) ExportPDF(q dto.ReportFilterQuery) ([]byte, string, error) {
	rows, err := s.subscriptionReportRows(q)
	if err != nil {
		return nil, "", err
	}

	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, "Laporan Langganan ENM Property", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)

	widths := []float64{45, 55, 45, 25, 30, 40, 45}
	headers := []string{"Kode Order", "Mitra", "Paket", "Jenis", "Status", "Jumlah", "Dibuat"}
	for i, header := range headers {
		pdf.CellFormat(widths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	for _, row := range rows {
		values := []string{
			row.OrderCode, row.PartnerName, row.PlanName, row.Kind, row.StatusLabel,
			row.AmountLabel, row.CreatedAt.Format("2006-01-02"),
		}
		for i, value := range values {
			pdf.CellFormat(widths[i], 7, value, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}

	buffer := new(bytes.Buffer)
	if err := pdf.Output(buffer); err != nil {
		return nil, "", httpx.Internal(err)
	}
	return buffer.Bytes(), "laporan-langganan.pdf", nil
}

// partnerStats menghitung mitra per status.
func (s *ReportService) partnerStats() (dto.PartnerStats, error) {
	partners, _, err := s.deps.Repos.Partners.List(database.ListQuery{PerPage: 0})
	if err != nil {
		return dto.PartnerStats{}, httpx.Internal(err)
	}

	stats := dto.PartnerStats{Total: len(partners)}
	for _, partner := range partners {
		switch partner.Status {
		case models.PartnerStatusActive:
			stats.Active++
		case models.PartnerStatusPending:
			stats.Pending++
		case models.PartnerStatusSuspended:
			stats.Suspended++
		case models.PartnerStatusExpired:
			stats.Expired++
		}
	}
	return stats, nil
}

// propertyStats menghitung properti per status.
func (s *ReportService) propertyStats() (dto.PropertyStats, error) {
	properties, _, err := s.deps.Repos.Properties.List(database.ListQuery{PerPage: 0})
	if err != nil {
		return dto.PropertyStats{}, httpx.Internal(err)
	}
	return countPropertyStatuses(properties), nil
}

// subscriptionStats menghitung langganan per status beserta yang hampir berakhir.
func (s *ReportService) subscriptionStats() (dto.SubscriptionStats, error) {
	subscriptions, _, err := s.deps.Repos.Subscriptions.List(database.ListQuery{PerPage: 0})
	if err != nil {
		return dto.SubscriptionStats{}, httpx.Internal(err)
	}

	now := timeNow()
	threshold := s.deps.Config.App.SubscriptionReminderDays
	stats := dto.SubscriptionStats{Total: len(subscriptions)}
	for _, subscription := range subscriptions {
		switch subscription.Status {
		case models.SubscriptionStatusActive:
			stats.Active++
			if subscription.DaysRemaining(now) <= threshold {
				stats.ExpiringSoon++
			}
		case models.SubscriptionStatusExpired:
			stats.Expired++
		case models.SubscriptionStatusCancelled:
			stats.Cancelled++
		}
	}
	return stats, nil
}

// orderStats menghitung order per status beserta pendapatannya.
func (s *ReportService) orderStats() (dto.OrderStats, error) {
	orders, _, err := s.deps.Repos.Orders.List(database.ListQuery{PerPage: 0})
	if err != nil {
		return dto.OrderStats{}, httpx.Internal(err)
	}

	now := timeNow()
	stats := dto.OrderStats{Total: len(orders)}
	for _, order := range orders {
		switch order.Status {
		case models.OrderStatusPending:
			stats.Pending++
		case models.OrderStatusPaid:
			stats.Paid++
			stats.Revenue += order.Amount
			if order.PaidAt != nil && sameMonth(*order.PaidAt, now) {
				stats.RevenueThisMonth += order.Amount
			}
		case models.OrderStatusFailed:
			stats.Failed++
		case models.OrderStatusExpired:
			stats.Expired++
		case models.OrderStatusCancelled:
			stats.Cancelled++
		}
	}
	stats.RevenueLabel = money.Rupiah(stats.Revenue)
	stats.RevenueThisMonthLabel = money.Rupiah(stats.RevenueThisMonth)
	return stats, nil
}

// subscriptionReportRows membentuk baris laporan langganan sesuai filter.
func (s *ReportService) subscriptionReportRows(q dto.ReportFilterQuery) ([]dto.SubscriptionReportRow, error) {
	filters := newFilters()
	putFilter(filters, "status", q.Status)
	putFilter(filters, "plan_id", q.PlanID)
	putFilter(filters, "partner_id", q.PartnerID)
	putFilter(filters, "from", q.From)
	putFilter(filters, "to", q.To)

	orders, _, err := s.deps.Repos.Orders.List(database.ListQuery{
		Search:  q.Search,
		PerPage: q.PerPage,
		Page:    q.Page,
		Filters: filters,
	})
	if err != nil {
		return nil, httpx.Internal(err)
	}

	rows := make([]dto.SubscriptionReportRow, 0, len(orders))
	for _, order := range orders {
		rows = append(rows, subscriptionRow(order))
	}
	return rows, nil
}

// partnerReportRows membentuk baris laporan mitra beserta aktivitas propertinya.
func (s *ReportService) partnerReportRows(q dto.ReportFilterQuery) ([]dto.PartnerReportRow, error) {
	filters := newFilters()
	putFilter(filters, "status", q.Status)
	putFilter(filters, "city", q.City)
	putFilter(filters, "province", q.Province)

	partners, _, err := s.deps.Repos.Partners.List(database.ListQuery{
		Search:  q.Search,
		PerPage: 0,
		Filters: filters,
	})
	if err != nil {
		return nil, httpx.Internal(err)
	}

	properties, _, err := s.deps.Repos.Properties.List(database.ListQuery{PerPage: 0})
	if err != nil {
		return nil, httpx.Internal(err)
	}
	propertiesByPartner := make(map[string][]*models.Property)
	for _, property := range properties {
		propertiesByPartner[property.PartnerID] = append(propertiesByPartner[property.PartnerID], property)
	}

	views, err := s.viewCount("")
	if err != nil {
		return nil, err
	}
	contacts, err := s.contactCount("")
	if err != nil {
		return nil, err
	}

	rows := make([]dto.PartnerReportRow, 0, len(partners))
	for _, partner := range partners {
		owned := propertiesByPartner[partner.ID]
		row := dto.PartnerReportRow{
			PartnerID:        partner.ID,
			BusinessName:     partner.BusinessName,
			PICName:          partner.PICName,
			Phone:            partner.Phone,
			City:             partner.City,
			Province:         partner.Province,
			Status:           string(partner.Status),
			StatusLabel:      partner.Status.Label(),
			PropertiesTotal:  len(owned),
			PropertyViews:    views[partner.ID],
			PropertyContacts: contacts[partner.ID],
			JoinedAt:         partner.JoinedAt,
		}
		for _, property := range owned {
			if property.Status == models.PropertyStatusActive {
				row.PropertiesActive++
			}
		}
		if subscription := s.subscriptions.Active(partner.ID); subscription != nil {
			row.PlanName = subscription.PlanName
			endsAt := subscription.EndsAt
			row.SubscriptionEndsAt = &endsAt
		}
		rows = append(rows, row)
	}

	database.SortByFunc(rows, func(a, b dto.PartnerReportRow) bool {
		return a.PropertyViews < b.PropertyViews
	}, true)
	return rows, nil
}

// viewCount menghitung tayangan properti, dikelompokkan per mitra. Mitra kosong
// berarti menghitung seluruh tayangan.
func (s *ReportService) viewCount(partnerID string) (map[string]int, error) {
	counts := make(map[string]int)
	for _, view := range s.deps.Repos.Analytics.Views() {
		if partnerID != "" && view.PartnerID != partnerID {
			continue
		}
		counts[view.PartnerID]++
	}
	return counts, nil
}

// contactCount menghitung lead properti, dikelompokkan per mitra.
func (s *ReportService) contactCount(partnerID string) (map[string]int, error) {
	counts := make(map[string]int)
	for _, contact := range s.deps.Repos.Analytics.Contacts() {
		if partnerID != "" && contact.PartnerID != partnerID {
			continue
		}
		counts[contact.PartnerID]++
	}
	return counts, nil
}

// recentProperties mengambil ringkasan properti terbaru sejumlah limit.
func (s *ReportService) recentProperties(properties []*models.Property, limit int) []dto.PropertySummary {
	if limit > len(properties) {
		limit = len(properties)
	}
	return dto.NewPropertySummaries(properties[:limit])
}

// countPropertyStatuses menghitung properti per status.
func countPropertyStatuses(properties []*models.Property) dto.PropertyStats {
	stats := dto.PropertyStats{Total: len(properties)}
	for _, property := range properties {
		switch property.Status {
		case models.PropertyStatusActive:
			stats.Active++
		case models.PropertyStatusDraft:
			stats.Draft++
		case models.PropertyStatusPendingReview:
			stats.PendingReview++
		case models.PropertyStatusInactive:
			stats.Inactive++
		case models.PropertyStatusExpired:
			stats.Expired++
		}
	}
	return stats
}

// subscriptionRow membentuk satu baris laporan dari order.
func subscriptionRow(order *models.Order) dto.SubscriptionReportRow {
	return dto.SubscriptionReportRow{
		OrderCode:   order.Code,
		PartnerID:   order.PartnerID,
		PartnerName: order.PartnerName,
		PlanID:      order.PlanID,
		PlanName:    order.PlanName,
		Kind:        string(order.Kind),
		Status:      string(order.Status),
		StatusLabel: order.Status.Label(),
		Provider:    string(order.Provider),
		Amount:      order.Amount,
		AmountLabel: money.Rupiah(order.Amount),
		CreatedAt:   order.CreatedAt,
		PaidAt:      order.PaidAt,
		ExpiresAt:   order.ExpiresAt,
	}
}

// sameMonth memeriksa apakah dua waktu berada pada bulan dan tahun yang sama.
func sameMonth(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month()
}

// formatTimePtr memformat waktu opsional untuk sel Excel.
func formatTimePtr(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04")
}
