-- 0020_agreement_checklist
--
-- Menambahkan poin ceklis persetujuan pada versi perjanjian kemitraan.
-- Setiap poin wajib dicentang mitra satu per satu saat menyetujui perjanjian
-- (mis. saat pertama masuk dashboard mitra atau aktivasi paket Gratis).

ALTER TABLE agreements
    ADD COLUMN IF NOT EXISTS checklist jsonb NOT NULL DEFAULT '[]'::jsonb;
