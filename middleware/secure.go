package middleware

import (
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

// SecureHeaders memasang header keamanan dasar: HSTS, nosniff, penolakan frame,
// dan Content Security Policy minimal.
func SecureHeaders() gin.HandlerFunc {
	return secure.New(secure.Config{
		STSSeconds:            31536000,
		STSIncludeSubdomains:  true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'",
		IENoOpen:              true,
	})
}
