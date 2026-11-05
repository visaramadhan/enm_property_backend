-- Kolom instruksi pembayaran (QRIS/VA) hasil Core API Midtrans.
-- Dipakai halaman pembayaran mitra untuk menampilkan QR atau nomor virtual account.
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS payment_method     text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payment_channel    text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payment_number     text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payment_qr_url     text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payment_qr_string  text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payment_expires_at timestamptz;
