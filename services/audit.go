package services

import (
	"bytes"
	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"

	"github.com/go-pdf/fpdf"
	"github.com/xuri/excelize/v2"
)

// AuditService mencatat dan membaca audit log aksi pengelola platform (BE-15).
type AuditService struct {
	deps Deps
}

// Log menyimpan satu catatan audit yang sudah terbentuk.
func (s *AuditService) Log(entry *models.AuditLog) error {
	if entry == nil {
		return nil
	}
	if entry.ID == "" {
		entry.ID = newID()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = timeNow()
	}
	return s.deps.Repos.Audits.Create(entry)
}

// Record membentuk catatan audit dari aksi pengelola platform. Kesalahan
// penyimpanan diabaikan agar tidak menggagalkan permintaan utama.
func (s *AuditService) Record(actor *models.User, action models.AuditAction, entity, entityID string, before, after any, note string, meta RequestMeta) {
	entry := &models.AuditLog{
		ID:        newID(),
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Before:    toJSON(before),
		After:     toJSON(after),
		Note:      note,
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
		CreatedAt: timeNow(),
	}
	if actor != nil {
		entry.ActorID = actor.ID
		entry.ActorName = actor.DisplayName()
		entry.ActorEmail = actor.Email
		entry.ActorRole = actor.Role
	}
	_ = s.Log(entry)
}

// List mengembalikan daftar audit log terfilter untuk halaman admin.
func (s *AuditService) List(q dto.AuditFilterQuery) (dto.ListResult[dto.AuditView], error) {
	entries, total, err := s.deps.Repos.Audits.List(auditListQuery(q))
	if err != nil {
		return dto.ListResult[dto.AuditView]{}, httpx.Internal(err)
	}
	return dto.NewListResult(dto.NewAuditViews(entries), total), nil
}

// Export mengembalikan seluruh audit log sesuai filter untuk diekspor.
func (s *AuditService) Export(q dto.AuditFilterQuery) ([]dto.AuditView, error) {
	query := auditListQuery(q)
	query.Page = 0
	query.PerPage = 0
	entries, _, err := s.deps.Repos.Audits.List(query)
	if err != nil {
		return nil, httpx.Internal(err)
	}
	return dto.NewAuditViews(entries), nil
}

// ExportXLSX membentuk berkas Excel dari audit log sesuai filter.
func (s *AuditService) ExportXLSX(q dto.AuditFilterQuery) ([]byte, string, error) {
	views, err := s.Export(q)
	if err != nil {
		return nil, "", err
	}

	file := excelize.NewFile()
	defer file.Close()
	const sheet = "Audit"
	index, err := file.NewSheet(sheet)
	if err != nil {
		return nil, "", httpx.Internal(err)
	}
	file.SetActiveSheet(index)
	_ = file.DeleteSheet("Sheet1")

	headers := []string{"Waktu", "Aktor", "Email", "Peran", "Aksi", "Entitas", "ID Entitas", "Catatan", "IP"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = file.SetCellValue(sheet, cell, header)
	}
	for r, view := range views {
		values := []any{
			view.CreatedAt.Format("2006-01-02 15:04"), view.ActorName, view.ActorEmail,
			view.ActorRole, view.Action, view.Entity, view.EntityID, view.Note, view.IPAddress,
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
	return buffer.Bytes(), "audit-log.xlsx", nil
}

// ExportPDF membentuk berkas PDF dari audit log sesuai filter.
func (s *AuditService) ExportPDF(q dto.AuditFilterQuery) ([]byte, string, error) {
	views, err := s.Export(q)
	if err != nil {
		return nil, "", err
	}

	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, "Audit Log ENM Property", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)

	widths := []float64{35, 45, 30, 30, 40, 50, 35}
	headers := []string{"Waktu", "Aktor", "Peran", "Aksi", "Entitas", "ID Entitas", "IP"}
	for i, header := range headers {
		pdf.CellFormat(widths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	for _, view := range views {
		values := []string{
			view.CreatedAt.Format("2006-01-02 15:04"), view.ActorName, view.ActorRole,
			view.Action, view.Entity, view.EntityID, view.IPAddress,
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
	return buffer.Bytes(), "audit-log.pdf", nil
}

// auditListQuery memetakan filter permintaan menjadi ListQuery repository.
func auditListQuery(q dto.AuditFilterQuery) database.ListQuery {
	filters := newFilters()
	putFilter(filters, "entity", q.Entity)
	putFilter(filters, "entity_id", q.EntityID)
	putFilter(filters, "actor_id", q.ActorID)
	putFilter(filters, "actor_role", q.ActorRole)
	putFilter(filters, "action", q.Action)
	putFilter(filters, "from", q.From)
	putFilter(filters, "to", q.To)
	return database.ListQuery{
		Search:  q.Search,
		Page:    q.Page,
		PerPage: q.PerPage,
		Filters: filters,
	}
}
