// Package money memformat nilai rupiah untuk laporan, invoice, dan ekspor.
package money

import (
	"strconv"
	"strings"
)

// Rupiah memformat nilai menjadi teks rupiah tanpa desimal (contoh Rp1.500.000).
func Rupiah(amount int64) string {
	if amount < 0 {
		return "-Rp" + grouped(strconv.FormatInt(-amount, 10))
	}
	return "Rp" + grouped(strconv.FormatInt(amount, 10))
}

// Ribuan memformat nilai dengan pemisah titik tanpa prefiks mata uang.
func Ribuan(amount int64) string {
	if amount < 0 {
		return "-" + grouped(strconv.FormatInt(-amount, 10))
	}
	return grouped(strconv.FormatInt(amount, 10))
}

// grouped menyisipkan titik sebagai pemisah ribuan setiap tiga digit.
func grouped(digits string) string {
	var builder strings.Builder
	for i, r := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			builder.WriteByte('.')
		}
		builder.WriteRune(r)
	}
	return builder.String()
}
