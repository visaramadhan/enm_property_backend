package middleware

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"enmproperty/backend/platform/httpx"
)

// validate dipakai bersama seluruh DTO dengan tag `validate`.
//
// Nama field pada pesan error mengikuti tag json agar sesuai kontrak API.
var validate = newValidator()

func newValidator() *validator.Validate {
	instance := validator.New(validator.WithRequiredStructEnabled())
	instance.SetTagName("validate")
	instance.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})
	return instance
}

// BindJSON memvalidasi body JSON.
func BindJSON(c *gin.Context, target any) bool {
	return bind(c, target, c.ShouldBindJSON)
}

// BindQuery memvalidasi parameter query.
func BindQuery(c *gin.Context, target any) bool {
	return bind(c, target, c.ShouldBindQuery)
}

// BindURI memvalidasi parameter path.
func BindURI(c *gin.Context, target any) bool {
	return bind(c, target, c.ShouldBindUri)
}

// Validate memvalidasi struct yang sudah terisi tanpa binding ulang.
func Validate(c *gin.Context, target any) bool {
	if err := validate.Struct(target); err != nil {
		respondBindingError(c, err)
		return false
	}
	return true
}

type binder func(target any) error

func bind(c *gin.Context, target any, fn binder) bool {
	if err := fn(target); err != nil {
		respondBindingError(c, err)
		return false
	}
	if err := validate.Struct(target); err != nil {
		respondBindingError(c, err)
		return false
	}
	return true
}

// respondBindingError mengubah kegagalan binding menjadi respons error standar.
func respondBindingError(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		httpx.Respond(c, httpx.Validation(validationDetails(validationErrors)))
		return
	}

	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		httpx.Respond(c, httpx.New(http.StatusRequestEntityTooLarge, httpx.CodeBadRequest,
			"Ukuran data yang dikirim melebihi batas"))
		return
	}

	httpx.Respond(c, httpx.BadRequest("Format permintaan tidak valid"))
}

// validationDetails menyusun pesan per field dalam bahasa Indonesia.
func validationDetails(errs validator.ValidationErrors) map[string]string {
	details := make(map[string]string, len(errs))
	for _, fieldError := range errs {
		details[fieldError.Field()] = validationMessage(fieldError)
	}
	return details
}

func validationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "url":
		return "format tautan tidak valid"
	case "uuid":
		return "format ID tidak valid"
	case "e164":
		return "format nomor telepon tidak valid"
	case "min":
		return "minimal " + fieldError.Param()
	case "max":
		return "maksimal " + fieldError.Param()
	case "gte":
		return "harus lebih besar atau sama dengan " + fieldError.Param()
	case "lte":
		return "harus lebih kecil atau sama dengan " + fieldError.Param()
	case "gt":
		return "harus lebih besar dari " + fieldError.Param()
	case "oneof":
		return "harus salah satu dari " + fieldError.Param()
	case "numeric":
		return "harus berupa angka"
	default:
		return "tidak valid (" + fieldError.Tag() + ")"
	}
}

// QueryInt mengambil nilai integer dari query dengan nilai bawaan.
func QueryInt(c *gin.Context, key string, fallback int) int {
	raw := c.Query(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
