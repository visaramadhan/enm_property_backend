package httpx

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body adalah amplop respons standar.
type Body struct {
	Success bool          `json:"success"`
	Data    any           `json:"data,omitempty"`
	Meta    *Meta         `json:"meta,omitempty"`
	Error   *ErrorPayload `json:"error,omitempty"`
}

// Meta memuat informasi paginasi.
type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ErrorPayload mengikuti format {code, message, details}.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// OK mengirim respons 200 dengan data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Success: true, Data: data})
}

// Created mengirim respons 201 dengan data.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Body{Success: true, Data: data})
}

// Accepted mengirim respons 202 dengan data.
func Accepted(c *gin.Context, data any) {
	c.JSON(http.StatusAccepted, Body{Success: true, Data: data})
}

// NoContent mengirim respons 204 tanpa body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Paginated mengirim respons daftar beserta metadata paginasi.
func Paginated(c *gin.Context, data any, page, perPage, total int) {
	totalPages := 1
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
		if totalPages == 0 {
			totalPages = 1
		}
	}
	c.JSON(http.StatusOK, Body{
		Success: true,
		Data:    data,
		Meta:    &Meta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages},
	})
}

// Fail mengirim respons gagal dengan kode dan pesan tertentu.
func Fail(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, Body{
		Success: false,
		Error:   &ErrorPayload{Code: code, Message: message},
	})
}

// Respond memetakan error menjadi respons gagal. Error yang bukan AppError
// dicatat sebagai 500 tanpa membocorkan detail internal ke klien.
func Respond(c *gin.Context, err error) {
	if err == nil {
		return
	}
	appErr := AsAppError(err)
	if appErr.Status >= http.StatusInternalServerError {
		slog.ErrorContext(c.Request.Context(), "permintaan gagal",
			"code", appErr.Code, "message", appErr.Message, "cause", appErr.Unwrap(),
		)
	}
	c.AbortWithStatusJSON(appErr.Status, Body{
		Success: false,
		Error: &ErrorPayload{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	})
}

// Abort menghentikan rantai handler dengan respons gagal.
func Abort(c *gin.Context, status int, code, message string) {
	c.Abort()
	Fail(c, status, code, message)
}

// Error mengambil error yang disimpan handler pada konteks, bila ada.
func Error(c *gin.Context) error {
	if value, exists := c.Get(ContextErrorKey); exists {
		if err, ok := value.(error); ok {
			return err
		}
	}
	return nil
}

// SetError menyimpan error pada konteks agar middleware ErrorHandler menanggapinya.
func SetError(c *gin.Context, err error) {
	c.Set(ContextErrorKey, err)
}

// ContextErrorKey adalah kunci konteks untuk menitipkan error ke middleware.
const ContextErrorKey = "httpx_error"
