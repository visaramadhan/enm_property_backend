// Package httpx menyeragamkan format respons dan error seluruh endpoint.
//
// Bentuk respons sukses: {"success":true,"data":...,"meta":...}
// Bentuk respons gagal:  {"success":false,"error":{"code":"...","message":"...","details":...}}
package httpx

import (
	"errors"
	"net/http"
)

// Kode error standar yang dipakai frontend untuk menentukan tindakan lanjutan.
const (
	CodeBadRequest     = "bad_request"
	CodeValidation     = "validation_error"
	CodeUnauthorized   = "unauthorized"
	CodeForbidden      = "forbidden"
	CodeNotFound       = "not_found"
	CodeConflict       = "conflict"
	CodeUnprocessable  = "unprocessable_entity"
	CodeTooManyRequest = "too_many_requests"
	CodeInternal       = "internal_error"
)

// AppError adalah error aplikasi yang membawa status HTTP, kode, dan pesan aman
// untuk ditampilkan ke klien. Detail internal tidak pernah ikut terkirim.
type AppError struct {
	Status  int
	Code    string
	Message string
	Details any
	cause   error
}

// Error memenuhi kontrak error.
func (e *AppError) Error() string { return e.Message }

// Unwrap mengembalikan penyebab asli untuk keperluan logging internal.
func (e *AppError) Unwrap() error { return e.cause }

// WithDetails melampirkan detail tambahan, umumnya peta field untuk validasi.
func (e *AppError) WithDetails(details any) *AppError {
	clone := *e
	clone.Details = details
	return &clone
}

// WithCause melampirkan error asli tanpa mengubah pesan yang dikirim ke klien.
func (e *AppError) WithCause(cause error) *AppError {
	clone := *e
	clone.cause = cause
	return &clone
}

// New membuat AppError baru.
func New(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

// BadRequest menandakan permintaan tidak dapat diproses.
func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, CodeBadRequest, message)
}

// Validation menandakan kegagalan validasi beserta detail per field.
func Validation(details any) *AppError {
	return New(http.StatusUnprocessableEntity, CodeValidation, "Data yang dikirim tidak valid").WithDetails(details)
}

// Unauthorized menandakan pengguna belum terautentikasi atau token tidak berlaku.
func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden menandakan pengguna tidak berhak mengakses sumber daya.
func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, CodeForbidden, message)
}

// NotFound menandakan sumber daya tidak ditemukan.
func NotFound(message string) *AppError {
	return New(http.StatusNotFound, CodeNotFound, message)
}

// Conflict menandakan benturan data, misalnya slug atau email yang sudah dipakai.
func Conflict(message string) *AppError {
	return New(http.StatusConflict, CodeConflict, message)
}

// Unprocessable menandakan permintaan valid secara format namun melanggar aturan bisnis.
func Unprocessable(message string) *AppError {
	return New(http.StatusUnprocessableEntity, CodeUnprocessable, message)
}

// TooManyRequests menandakan pembatasan laju permintaan.
func TooManyRequests(message string) *AppError {
	return New(http.StatusTooManyRequests, CodeTooManyRequest, message)
}

// Internal membungkus error tak terduga menjadi respons 500 yang aman.
func Internal(cause error) *AppError {
	return New(http.StatusInternalServerError, CodeInternal, "Terjadi kesalahan pada server").WithCause(cause)
}

// AsAppError mengambil AppError dari rantai error, atau membungkusnya sebagai 500.
func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}
