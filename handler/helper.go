// Package handler memuat seluruh HTTP handler yang menerjemahkan permintaan
// klien menjadi pemanggilan service, lalu menyusun respons standar.
//
// Handler tidak memuat logika bisnis: validasi bentuk data dilakukan melalui
// middleware.ValidateRequest, sedangkan aturan bisnis berada di lapisan service.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// requestMeta mengumpulkan jejak permintaan untuk bukti persetujuan perjanjian
// dan pencatatan audit manual.
func requestMeta(c *gin.Context) services.RequestMeta {
	return services.RequestMeta{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}

// pageParams mengambil nomor halaman dan ukuran halaman dari query string
// dengan nilai bawaan yang aman.
func pageParams(c *gin.Context) (int, int) {
	page := middleware.QueryInt(c, "page", 1)
	if page < 1 {
		page = 1
	}
	perPage := middleware.QueryInt(c, "per_page", 12)
	if perPage <= 0 {
		perPage = 12
	}
	return page, perPage
}

// paginate mengirim daftar beserta metadata paginasi standar.
func paginate[T any](c *gin.Context, result dto.ListResult[T], page, perPage int) {
	httpx.Paginated(c, result.Items, page, perPage, result.Total)
}

// download mengirim berkas hasil ekspor sebagai lampiran unduhan.
func download(c *gin.Context, data []byte, contentType, filename string) {
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, contentType, data)
}

// idParam mengambil parameter path bernama id tanpa validasi tambahan.
func idParam(c *gin.Context, name string) string {
	return c.Param(name)
}
