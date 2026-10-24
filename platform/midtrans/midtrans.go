// Package midtrans membungkus Midtrans Snap dan Core API untuk checkout paket,
// pengecekan status transaksi, dan verifikasi signature webhook.
//
// Bila MIDTRANS_ENABLED=false atau server key belum diisi, klien berjalan pada
// mode simulasi: token dibuat lokal dan transaksi dianggap langsung lunas
// sehingga alur checkout paket tetap dapat diuji tanpa akun Midtrans.
package midtrans

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"

	"enmproperty/backend/config"
)

// Item adalah satu baris rincian pembayaran pada Snap.
type Item struct {
	ID       string
	Name     string
	Price    int64
	Quantity int32
}

// Request adalah permintaan pembuatan token Snap.
type Request struct {
	OrderID       string
	Amount        int64
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
	Items         []Item
	ExpiryMinutes int
}

// SnapResult adalah hasil pembuatan token Snap.
type SnapResult struct {
	Token       string
	RedirectURL string
	Simulated   bool
}

// StatusResult adalah ringkasan status transaksi dari Midtrans.
type StatusResult struct {
	OrderID           string
	TransactionID     string
	TransactionStatus string
	FraudStatus       string
	PaymentType       string
	GrossAmount       string
	StatusCode        string
}

// Client membungkus klien Snap dan Core Midtrans.
type Client struct {
	serverKey  string
	simulate   bool
	snapClient snap.Client
	coreClient coreapi.Client
}

// New membuat klien Midtrans sesuai konfigurasi.
func New(cfg config.Midtrans) *Client {
	environment := midtrans.Sandbox
	if cfg.IsProduction {
		environment = midtrans.Production
	}

	client := &Client{
		serverKey: cfg.ServerKey,
		simulate:  !cfg.Enabled || cfg.ServerKey == "",
	}
	if !client.simulate {
		client.snapClient.New(cfg.ServerKey, environment)
		client.coreClient.New(cfg.ServerKey, environment)
	}
	return client
}

// Simulated melaporkan apakah klien berjalan tanpa Midtrans sungguhan.
func (c *Client) Simulated() bool { return c.simulate }

// CreateSnapToken membuat token Snap untuk satu order.
func (c *Client) CreateSnapToken(_ context.Context, req Request) (*SnapResult, error) {
	if c.simulate {
		return &SnapResult{Token: "simulate-" + req.OrderID, Simulated: true}, nil
	}

	items := make([]midtrans.ItemDetails, 0, len(req.Items))
	for _, item := range req.Items {
		quantity := item.Quantity
		if quantity == 0 {
			quantity = 1
		}
		items = append(items, midtrans.ItemDetails{
			ID:    item.ID,
			Name:  item.Name,
			Price: item.Price,
			Qty:   quantity,
		})
	}
	if len(items) == 0 {
		items = append(items, midtrans.ItemDetails{
			ID:    req.OrderID,
			Name:  "Langganan ENM Property",
			Price: req.Amount,
			Qty:   1,
		})
	}

	request := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  req.OrderID,
			GrossAmt: req.Amount,
		},
		Items: &items,
		CustomerDetail: &midtrans.CustomerDetails{
			FName: req.CustomerName,
			Email: req.CustomerEmail,
			Phone: req.CustomerPhone,
		},
	}
	if req.ExpiryMinutes > 0 {
		request.Expiry = &snap.ExpiryDetails{
			Unit:     "minute",
			Duration: int64(req.ExpiryMinutes),
		}
	}

	response, mErr := c.snapClient.CreateTransaction(request)
	if mErr != nil {
		return nil, fmt.Errorf("gagal membuat token Midtrans: %w", mErr)
	}
	return &SnapResult{Token: response.Token, RedirectURL: response.RedirectURL}, nil
}

// TransactionStatus mengambil status transaksi terkini dari Midtrans.
func (c *Client) TransactionStatus(_ context.Context, orderID string) (*StatusResult, error) {
	if c.simulate {
		return &StatusResult{
			OrderID:           orderID,
			TransactionStatus: "settlement",
			StatusCode:        "200",
		}, nil
	}

	response, mErr := c.coreClient.CheckTransaction(orderID)
	if mErr != nil {
		return nil, fmt.Errorf("gagal memeriksa status transaksi: %w", mErr)
	}
	return &StatusResult{
		OrderID:           response.OrderID,
		TransactionID:     response.TransactionID,
		TransactionStatus: response.TransactionStatus,
		FraudStatus:       response.FraudStatus,
		PaymentType:       response.PaymentType,
		GrossAmount:       response.GrossAmount,
		StatusCode:        response.StatusCode,
	}, nil
}

// VerifySignature memeriksa signature_key webhook Midtrans.
//
// Rumus Midtrans: SHA512(order_id + status_code + gross_amount + server_key).
// Pada mode simulasi selalu true karena webhook tidak dipakai.
func (c *Client) VerifySignature(orderID, statusCode, grossAmount, signature string) bool {
	if c.simulate {
		return true
	}
	if signature == "" {
		return false
	}
	payload := orderID + statusCode + grossAmount + c.serverKey
	sum := sha512.Sum512([]byte(payload))
	return strings.EqualFold(hex.EncodeToString(sum[:]), signature)
}
