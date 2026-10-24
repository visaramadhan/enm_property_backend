// Package models berisi seluruh entitas domain ENMProperty beserta tipe
// enumerasi yang dipakai bersama oleh services dan handler.
package models

// Role adalah peran pengguna yang juga disimpan sebagai custom claim Firebase.
type Role string

// Daftar peran yang dikenal sistem.
const (
	RoleCustomer   Role = "customer"
	RoleMitra      Role = "mitra"
	RoleAdmin      Role = "admin"
	RoleSuperadmin Role = "superadmin"
	RoleOwner      Role = "owner"
)

// Valid memeriksa apakah role termasuk peran yang dikenal.
func (r Role) Valid() bool {
	switch r {
	case RoleCustomer, RoleMitra, RoleAdmin, RoleSuperadmin, RoleOwner:
		return true
	default:
		return false
	}
}

// IsInternal mengembalikan true untuk peran pengelola platform.
func (r Role) IsInternal() bool {
	return r == RoleAdmin || r == RoleSuperadmin || r == RoleOwner
}

// Label mengembalikan nama peran dalam bahasa Indonesia.
func (r Role) Label() string {
	switch r {
	case RoleCustomer:
		return "Customer"
	case RoleMitra:
		return "Mitra"
	case RoleAdmin:
		return "Admin"
	case RoleSuperadmin:
		return "Superadmin"
	case RoleOwner:
		return "Owner"
	default:
		return string(r)
	}
}

// UserStatus adalah status aktif akun.
type UserStatus string

// Daftar status akun.
const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

// PartnerStatus adalah status kemitraan.
type PartnerStatus string

// Daftar status kemitraan.
const (
	PartnerStatusPending   PartnerStatus = "pending"
	PartnerStatusActive    PartnerStatus = "active"
	PartnerStatusSuspended PartnerStatus = "suspended"
	PartnerStatusExpired   PartnerStatus = "expired"
)

// Label mengembalikan status kemitraan dalam bahasa Indonesia.
func (s PartnerStatus) Label() string {
	switch s {
	case PartnerStatusPending:
		return "Menunggu verifikasi"
	case PartnerStatusActive:
		return "Aktif"
	case PartnerStatusSuspended:
		return "Nonaktif"
	case PartnerStatusExpired:
		return "Kedaluwarsa"
	default:
		return string(s)
	}
}

// SubscriptionStatus adalah status masa aktif langganan.
type SubscriptionStatus string

// Daftar status langganan.
const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
)

// OrderStatus adalah status transaksi pembayaran paket.
type OrderStatus string

// Daftar status order.
const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusExpired   OrderStatus = "expired"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Label mengembalikan status order dalam bahasa Indonesia untuk laporan.
func (s OrderStatus) Label() string {
	switch s {
	case OrderStatusPending:
		return "Menunggu"
	case OrderStatusPaid:
		return "Berhasil"
	case OrderStatusFailed:
		return "Gagal"
	case OrderStatusExpired:
		return "Kedaluwarsa"
	case OrderStatusCancelled:
		return "Dibatalkan"
	default:
		return string(s)
	}
}

// OrderKind membedakan checkout baru, perpanjangan, upgrade, dan paket gratis.
type OrderKind string

// Daftar jenis order.
const (
	OrderKindNew     OrderKind = "new"
	OrderKindRenewal OrderKind = "renewal"
	OrderKindUpgrade OrderKind = "upgrade"
	OrderKindFree    OrderKind = "free"
)

// PaymentProvider adalah penyedia pembayaran order.
type PaymentProvider string

// Daftar penyedia pembayaran.
const (
	PaymentProviderMidtrans  PaymentProvider = "midtrans"
	PaymentProviderFree      PaymentProvider = "free"
	PaymentProviderSimulated PaymentProvider = "simulated"
)

// PropertyStatus adalah status tayang properti.
type PropertyStatus string

// Daftar status properti.
const (
	PropertyStatusDraft         PropertyStatus = "draft"
	PropertyStatusPendingReview PropertyStatus = "pending_review"
	PropertyStatusActive        PropertyStatus = "active"
	PropertyStatusInactive      PropertyStatus = "inactive"
	PropertyStatusExpired       PropertyStatus = "expired"
)

// Label mengembalikan status properti dalam bahasa Indonesia.
func (s PropertyStatus) Label() string {
	switch s {
	case PropertyStatusDraft:
		return "Draft"
	case PropertyStatusPendingReview:
		return "Menunggu review"
	case PropertyStatusActive:
		return "Tayang"
	case PropertyStatusInactive:
		return "Nonaktif"
	case PropertyStatusExpired:
		return "Kedaluwarsa"
	default:
		return string(s)
	}
}

// ListingType membedakan properti dijual dan disewa.
type ListingType string

// Daftar jenis penawaran.
const (
	ListingTypeSale ListingType = "dijual"
	ListingTypeRent ListingType = "disewa"
)

// Valid memeriksa jenis penawaran yang dikenal.
func (l ListingType) Valid() bool {
	return l == ListingTypeSale || l == ListingTypeRent
}

// PricePeriod adalah satuan periode harga sewa.
type PricePeriod string

// Daftar periode harga.
const (
	PricePeriodDaily   PricePeriod = "hari"
	PricePeriodMonthly PricePeriod = "bulan"
	PricePeriodYearly  PricePeriod = "tahun"
)

// Valid memeriksa periode harga yang dikenal.
func (p PricePeriod) Valid() bool {
	return p == PricePeriodDaily || p == PricePeriodMonthly || p == PricePeriodYearly
}

// PropertyType adalah jenis properti.
type PropertyType string

// Daftar jenis properti.
const (
	PropertyTypeHouse     PropertyType = "rumah"
	PropertyTypeApartment PropertyType = "apartemen"
	PropertyTypeRuko      PropertyType = "ruko"
	PropertyTypeLand      PropertyType = "tanah"
	PropertyTypeVilla     PropertyType = "villa"
	PropertyTypeOffice    PropertyType = "kantor"
	PropertyTypeWarehouse PropertyType = "gudang"
	PropertyTypeOther     PropertyType = "lainnya"
)

// Valid memeriksa jenis properti yang dikenal.
func (p PropertyType) Valid() bool {
	switch p {
	case PropertyTypeHouse, PropertyTypeApartment, PropertyTypeRuko, PropertyTypeLand,
		PropertyTypeVilla, PropertyTypeOffice, PropertyTypeWarehouse, PropertyTypeOther:
		return true
	default:
		return false
	}
}

// AuditAction adalah aksi yang dicatat pada audit log.
type AuditAction string

// Daftar aksi audit.
const (
	AuditActionCreate        AuditAction = "create"
	AuditActionUpdate        AuditAction = "update"
	AuditActionDelete        AuditAction = "delete"
	AuditActionActivate      AuditAction = "activate"
	AuditActionDeactivate    AuditAction = "deactivate"
	AuditActionResetPassword AuditAction = "reset_password"
	AuditActionExport        AuditAction = "export"
	AuditActionLogin         AuditAction = "login"
)
