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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

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

// ChargeResult adalah ringkasan hasil pembuatan transaksi Core API (QRIS/VA).
//
// Pada mode simulasi seluruh nilai diisi secara deterministik agar halaman
// pembayaran tetap dapat diuji tanpa akun Midtrans.
type ChargeResult struct {
	TransactionID     string
	OrderID           string
	GrossAmount       string
	PaymentType       string
	QRString          string
	QRURL             string
	VANumber          string
	Bank              string
	ExpiryTime        string
	TransactionStatus string
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

// ErrorDetail mengambil kode status HTTP dan pesan dari respons Midtrans bila
// err berasal dari Midtrans. ok bernilai false untuk error non-Midtrans.
//
// Dipakai agar penyebab kegagalan charge (mis. kanal QRIS belum aktif) dapat
// diteruskan ke pengguna, bukan disamarkan menjadi 500 tanpa keterangan.
func ErrorDetail(err error) (status int, message string, ok bool) {
	var midtransErr *midtrans.Error
	if errors.As(err, &midtransErr) {
		return midtransErr.GetStatusCode(), midtransErr.GetMessage(), true
	}
	return 0, "", false
}

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

// ChargeQRIS membuat transaksi QRIS melalui Core API agar kode QR dapat
// dirender langsung pada halaman pembayaran.
func (c *Client) ChargeQRIS(_ context.Context, req Request) (*ChargeResult, error) {
	if c.simulate {
		return simulatedCharge("qris", "qris", req), nil
	}

	chargeReq := newChargeReq(req, coreapi.PaymentTypeQris)
	chargeReq.Qris = &coreapi.QrisDetails{}
	return c.charge(chargeReq)
}

// ChargeVA membuat transaksi bank transfer (virtual account) melalui Core API.
// bankName bernilai bca, bni, bri, atau permata.
func (c *Client) ChargeVA(_ context.Context, bankName string, req Request) (*ChargeResult, error) {
	bank := strings.ToLower(strings.TrimSpace(bankName))
	if c.simulate {
		return simulatedCharge("bank_transfer", bank, req), nil
	}

	chargeReq := newChargeReq(req, coreapi.PaymentTypeBankTransfer)
	chargeReq.BankTransfer = &coreapi.BankTransferDetails{Bank: midtrans.Bank(bank)}
	return c.charge(chargeReq)
}

// newChargeReq melengkapi permintaan charge dengan data transaksi, pelanggan,
// dan masa berlaku.
func newChargeReq(req Request, paymentType coreapi.CoreapiPaymentType) *coreapi.ChargeReq {
	chargeReq := &coreapi.ChargeReq{
		PaymentType: paymentType,
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  req.OrderID,
			GrossAmt: req.Amount,
		},
		CustomerDetails: &midtrans.CustomerDetails{
			FName: req.CustomerName,
			Email: req.CustomerEmail,
			Phone: req.CustomerPhone,
		},
	}
	if req.ExpiryMinutes > 0 {
		chargeReq.CustomExpiry = &coreapi.CustomExpiry{
			ExpiryDuration: req.ExpiryMinutes,
			Unit:           "minute",
		}
	}
	return chargeReq
}

// charge mengirim permintaan charge ke Core API dan meringkas responsnya.
func (c *Client) charge(req *coreapi.ChargeReq) (*ChargeResult, error) {
	response, mErr := c.coreClient.ChargeTransaction(req)
	if mErr != nil {
		return nil, fmt.Errorf("gagal membuat transaksi Midtrans: %w", mErr)
	}

	result := &ChargeResult{
		TransactionID:     response.TransactionID,
		OrderID:           response.OrderID,
		GrossAmount:       response.GrossAmount,
		PaymentType:       response.PaymentType,
		QRString:          response.QRString,
		VANumber:          response.PermataVaNumber,
		Bank:              response.Bank,
		ExpiryTime:        response.ExpiryTime,
		TransactionStatus: response.TransactionStatus,
	}
	if len(response.VaNumbers) > 0 {
		result.VANumber = response.VaNumbers[0].VANumber
		result.Bank = response.VaNumbers[0].Bank
	}
	for _, action := range response.Actions {
		if strings.HasPrefix(strings.ToLower(action.Name), "generate-qr") {
			result.QRURL = action.URL
			break
		}
	}
	return result, nil
}

// simulatedCharge membentuk data pembayaran palsu yang deterministik sehingga
// alur pembayaran tetap dapat diuji tanpa memanggil jaringan Midtrans.
func simulatedCharge(paymentType, method string, req Request) *ChargeResult {
	minutes := req.ExpiryMinutes
	if minutes <= 0 {
		minutes = 60
	}
	expiry := time.Now().Add(time.Duration(minutes) * time.Minute)

	result := &ChargeResult{
		TransactionID:     "simulate-" + req.OrderID,
		OrderID:           req.OrderID,
		GrossAmount:       strconv.FormatInt(req.Amount, 10),
		PaymentType:       paymentType,
		TransactionStatus: "pending",
		ExpiryTime:        expiry.Format("2006-01-02 15:04:05"),
	}
	if method == "qris" {
		result.QRString = "SIMULASI-QRIS:" + req.OrderID
	} else {
		result.Bank = method
		result.VANumber = simulatedVANumber(req.OrderID)
	}
	return result
}

// simulatedVANumber membentuk nomor VA palsu: awalan 8808 diikuti 10 digit
// terakhir kode order.
func simulatedVANumber(orderID string) string {
	digits := make([]byte, 0, len(orderID))
	for i := 0; i < len(orderID); i++ {
		if orderID[i] >= '0' && orderID[i] <= '9' {
			digits = append(digits, orderID[i])
		}
	}
	if len(digits) > 10 {
		digits = digits[len(digits)-10:]
	}
	for len(digits) < 10 {
		digits = append([]byte{'0'}, digits...)
	}
	return "8808" + string(digits)
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
