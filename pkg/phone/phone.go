// Package phone menormalkan nomor telepon Indonesia ke format WhatsApp 62...
package phone

import (
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// DefaultRegion adalah kode negara yang dipakai bila nomor tidak memuat awalan.
const DefaultRegion = "ID"

// Normalize mengubah nomor menjadi format E.164 tanpa tanda plus (contoh 6281234567890).
//
// Bila nomor tidak dapat diurai, teks asli (setelah dipangkas) dikembalikan agar
// alur penyimpanan tetap berjalan tanpa menggagalkan permintaan.
func Normalize(input string) string {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return ""
	}

	parsed, err := phonenumbers.Parse(raw, DefaultRegion)
	if err != nil {
		return raw
	}
	formatted := phonenumbers.Format(parsed, phonenumbers.E164)
	return strings.TrimPrefix(formatted, "+")
}

// IsValid memeriksa apakah nomor valid untuk region Indonesia.
func IsValid(input string) bool {
	parsed, err := phonenumbers.Parse(input, DefaultRegion)
	if err != nil {
		return false
	}
	return phonenumbers.IsValidNumber(parsed)
}

// WaMeURL membentuk tautan wa.me dari nomor yang sudah dinormalkan.
func WaMeURL(number, message string) string {
	normalized := Normalize(number)
	if normalized == "" {
		return ""
	}
	url := "https://wa.me/" + normalized
	if message != "" {
		url += "?text=" + message
	}
	return url
}
