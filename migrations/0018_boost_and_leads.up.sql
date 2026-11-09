-- 0018_boost_and_leads
--
-- Menambahkan dua hal:
--   1. Kuota boost per paket (plans.boost_quota) yang diatur Superadmin.
--   2. Detail lead pada property_contacts agar form kontak publik dapat
--      menyimpan nama, telepon, dan pesan calon pembeli.

ALTER TABLE plans
    ADD COLUMN IF NOT EXISTS boost_quota integer NOT NULL DEFAULT 0;

ALTER TABLE property_contacts
    ADD COLUMN IF NOT EXISTS contact_name text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS contact_phone text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS message text NOT NULL DEFAULT '';

-- Mempercepat rekap lead per properti maupun per mitra.
CREATE INDEX IF NOT EXISTS idx_property_contacts_partner ON property_contacts (partner_id, created_at DESC);
