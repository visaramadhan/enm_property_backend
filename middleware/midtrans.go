package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/platform/midtrans"
)

// midtransNotification adalah subset payload webhook yang dipakai verifikasi.
type midtransNotification struct {
	OrderID      string `json:"order_id"`
	StatusCode   string `json:"status_code"`
	GrossAmount  string `json:"gross_amount"`
	SignatureKey string `json:"signature_key"`
}

// MidtransSignature memverifikasi signature_key webhook sebelum diproses.
// Body dibaca ulang sehingga handler tetap dapat mengurai payload yang sama.
func MidtransSignature(client *midtrans.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := io.ReadAll(c.Request.Body)
		if err != nil {
			httpx.Abort(c, http.StatusBadRequest, httpx.CodeBadRequest, "Body notifikasi tidak dapat dibaca")
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(raw))

		var payload midtransNotification
		if err := json.Unmarshal(raw, &payload); err != nil {
			httpx.Abort(c, http.StatusBadRequest, httpx.CodeBadRequest, "Format notifikasi tidak valid")
			return
		}

		if !client.VerifySignature(payload.OrderID, payload.StatusCode, payload.GrossAmount, payload.SignatureKey) {
			httpx.Abort(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "Signature notifikasi tidak valid")
			return
		}

		c.Next()
	}
}
