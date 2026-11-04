package handler

import (
	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// InvoiceHandler melayani tampilan dan unduhan invoice langganan.
type InvoiceHandler struct {
	services *services.Services
}

// NewInvoiceHandler membuat handler invoice.
func NewInvoiceHandler(svc *services.Services) *InvoiceHandler {
	return &InvoiceHandler{services: svc}
}

// Get mengembalikan rincian invoice milik pengguna yang sedang masuk.
func (h *InvoiceHandler) Get(c *gin.Context) {
	var params dto.CodeParam
	if !middleware.BindURI(c, &params) {
		return
	}

	invoice, err := h.services.Payment.GetInvoice(middleware.CurrentUser(c), params.OrderCode)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, invoice)
}

// PDF mengirim invoice sebagai berkas PDF yang dapat diunduh.
func (h *InvoiceHandler) PDF(c *gin.Context) {
	var params dto.CodeParam
	if !middleware.BindURI(c, &params) {
		return
	}

	invoice, err := h.services.Payment.GetInvoice(middleware.CurrentUser(c), params.OrderCode)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	data, err := renderInvoicePDF(invoice)
	if err != nil {
		httpx.Respond(c, httpx.Internal(err))
		return
	}
	download(c, data, "application/pdf", "invoice-"+params.OrderCode+".pdf")
}

// renderInvoicePDF menyusun berkas PDF sederhana dari satu invoice.
func renderInvoicePDF(invoice dto.InvoiceView) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Invoice "+invoice.InvoiceNumber, "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 7, "Mitra: "+invoice.Partner.BusinessName, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 7, "Paket: "+invoice.Order.PlanName, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 7, "Status: "+invoice.Order.StatusLabel, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 7, "Tanggal: "+invoice.IssuedAt.Format("2006-01-02 15:04"), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(110, 8, "Item", "1", 0, "L", true, 0, "")
	pdf.CellFormat(40, 8, "Deskripsi", "1", 0, "L", true, 0, "")
	pdf.CellFormat(40, 8, "Total", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	for _, item := range invoice.Items {
		pdf.CellFormat(110, 7, item.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 7, item.Description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 7, item.TotalLabel, "1", 1, "R", false, 0, "")
	}

	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(150, 8, "Total", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, invoice.TotalLabel, "1", 1, "R", false, 0, "")

	buffer := new(bytes.Buffer)
	if err := pdf.Output(buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
