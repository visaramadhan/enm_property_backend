package middleware

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// Compress mengaktifkan kompresi gzip pada respons.
func Compress() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression)
}
