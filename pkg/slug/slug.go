// Package slug membuat slug URL-safe dari judul properti dan judul perjanjian.
package slug

import (
	"strconv"
	"strings"

	goslug "github.com/gosimple/slug"
)

// Make mengubah teks menjadi slug huruf kecil dengan pemisah tanda hubung.
func Make(text string) string {
	return goslug.Make(text)
}

// MakeWithSuffix menambahkan akhiran numerik untuk menjaga keunikan slug.
func MakeWithSuffix(text string, seq int) string {
	base := Make(text)
	if seq <= 1 {
		return base
	}
	return base + "-" + strconv.Itoa(seq)
}

// Normalize membersihkan slug yang dikirim klien agar konsisten.
func Normalize(text string) string {
	return strings.Trim(goslug.Make(text), "-")
}
