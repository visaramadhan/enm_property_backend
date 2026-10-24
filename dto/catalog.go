package dto

// ContactResponse adalah respons GET /properties/{id}/contact berisi nomor
// WhatsApp mitra, atau nomor admin sebagai cadangan.
type ContactResponse struct {
	PropertyID    string `json:"property_id"`
	PropertyTitle string `json:"property_title"`
	PartnerName   string `json:"partner_name"`
	Phone         string `json:"phone"`
	WhatsAppURL   string `json:"whatsapp_url"`
	Message       string `json:"message"`
	IsFallback    bool   `json:"is_fallback"`
}
