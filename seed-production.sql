-- =============================================================================
-- Seed data awal untuk PostgreSQL (DB_DRIVER=postgres)
--
-- Data ini meniru database.Seed() di backend yang hanya berjalan pada mode
-- memory. Jalankan SEKALI setelah migrasi pertama selesai:
--
--   docker compose exec -T postgres_db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"' < seed-production.sql
--
-- Klausa ON CONFLICT membuat skrip ini idempotent (aman dijalankan ulang).
-- =============================================================================

-- --- Perjanjian kemitraan aktif ------------------------------------------------
INSERT INTO agreements (id, version, title, content, is_active, effective_at, created_by)
VALUES (
  'agr-2026-01',
  '2026.01',
  'Perjanjian Kemitraan ENM Property',
  'Dengan mendaftar sebagai mitra, Anda menyetujui ketentuan pemuatan listing, kebenaran data properti, dan kebijakan moderasi platform.',
  true,
  now(),
  'system'
)
ON CONFLICT (id) DO UPDATE
  SET version = EXCLUDED.version,
      title = EXCLUDED.title,
      content = EXCLUDED.content,
      is_active = EXCLUDED.is_active;

-- --- Paket langganan -----------------------------------------------------------
INSERT INTO plans (id, slug, name, description, price, duration_days, listing_quota,
                   features, is_free, is_active, sort_order, level, boost_enabled,
                   boost_quota, keyword_access, catalog_access, whatsapp_access,
                   created_at, updated_at)
VALUES
  (
    'pln-free', 'gratis', 'Gratis',
    'Paket uji coba untuk mitra baru.',
    0, 30, 3,
    '["3 listing aktif","Masa aktif 30 hari","Dukungan email"]'::jsonb,
    true, true, 1, 1, false, 0, false, false, false,
    now(), now()
  ),
  (
    'pln-standar', 'standar', 'Standar',
    'Cocok untuk agen properti perorangan.',
    750000, 90, 15,
    '["15 listing aktif","Masa aktif 90 hari","Statistik tayangan","Akses pengaturan keyword","Akses pengaturan katalog","Akses pengaturan nomor WA","Dukungan WhatsApp"]'::jsonb,
    false, true, 2, 2, true, 2, true, true, true,
    now(), now()
  ),
  (
    'pln-premium', 'premium', 'Premium',
    'Untuk agensi dengan banyak listing.',
    1500000, 90, 25,
    '["25 listing aktif","Masa aktif 90 hari","Prioritas review","Boost listing level 3","Statistik lanjutan","Dukungan prioritas"]'::jsonb,
    false, true, 3, 3, true, 3, true, true, true,
    now(), now()
  )
ON CONFLICT (id) DO UPDATE
  SET name = EXCLUDED.name,
      description = EXCLUDED.description,
      price = EXCLUDED.price,
      is_active = EXCLUDED.is_active;

-- --- Kategori properti ---------------------------------------------------------
INSERT INTO categories (id, slug, name, description, sort_order, is_active, created_at, updated_at)
VALUES
  ('cat-rumah', 'rumah', 'Rumah', 'Rumah tinggal, cluster, dan hunian keluarga.', 1, true, now(), now()),
  ('cat-apartemen', 'apartemen', 'Apartemen', 'Unit apartemen dan kondominium.', 2, true, now(), now()),
  ('cat-tanah', 'tanah', 'Tanah', 'Kavling tanah dan lahan siap bangun.', 3, true, now(), now()),
  ('cat-ruko', 'ruko', 'Ruko', 'Rumah toko dan ruang usaha komersial.', 4, true, now(), now()),
  ('cat-villa', 'villa', 'Villa', 'Villa dan properti liburan.', 5, true, now(), now())
ON CONFLICT (id) DO UPDATE
  SET name = EXCLUDED.name,
      is_active = EXCLUDED.is_active;

-- --- Fasilitas properti --------------------------------------------------------
INSERT INTO facilities (id, slug, name, description, icon, sort_order, is_active, created_at, updated_at)
VALUES
  ('fac-carport', 'carport', 'Carport', 'Tempat parkir kendaraan di dalam properti.', 'car', 1, true, now(), now()),
  ('fac-kolam-renang', 'kolam-renang', 'Kolam Renang', 'Kolam renang pribadi atau bersama.', 'waves', 2, true, now(), now()),
  ('fac-taman', 'taman', 'Taman', 'Area taman hijau di dalam properti.', 'trees', 3, true, now(), now()),
  ('fac-keamanan-24-jam', 'keamanan-24-jam', 'Keamanan 24 Jam', 'Sistem keamanan yang beroperasi 24 jam.', 'shield-check', 4, true, now(), now()),
  ('fac-ac', 'ac', 'AC', 'Pendingin ruangan (air conditioner).', 'snowflake', 5, true, now(), now()),
  ('fac-dapur', 'dapur', 'Dapur', 'Dapur lengkap dengan peralatan masak.', 'cooking-pot', 6, true, now(), now()),
  ('fac-wifi', 'wifi', 'Wi-Fi', 'Akses internet nirkabel.', 'wifi', 7, true, now(), now()),
  ('fac-listrik', 'listrik', 'Listrik', 'Pasokan listrik tersedia.', 'zap', 8, true, now(), now()),
  ('fac-air', 'air', 'Air', 'Pasokan air bersih tersedia.', 'droplets', 9, true, now(), now()),
  ('fac-furnished', 'furnished', 'Furnished', 'Properti sudah dilengkapi furnitur.', 'sofa', 10, true, now(), now())
ON CONFLICT (id) DO UPDATE
  SET name = EXCLUDED.name,
      is_active = EXCLUDED.is_active;

-- --- Pengaturan sistem ---------------------------------------------------------
INSERT INTO settings (key, value, description, is_secret, updated_by, updated_at)
VALUES
  ('site_name', 'ENM Property', 'Nama situs yang tampil di header dan metadata.', false, 'system', now()),
  ('admin_whatsapp', '+6281100000000', 'Nomor WhatsApp admin untuk inquiry mitra.', false, 'system', now()),
  ('subscription_alert_days', '7', 'Ambang hari peringatan sebelum langganan berakhir.', false, 'system', now()),
  ('contact_fallback_whatsapp', '+6281100000000', 'Nomor fallback bila mitra belum mengisi kontak.', false, 'system', now()),
  ('agreement_required', 'true', 'Wajib menyetujui perjanjian sebelum mengunggah listing.', false, 'system', now()),
  ('site_logo_url', '', 'URL logo situs. Kosongkan untuk memakai monogram nama situs.', false, 'system', now()),
  ('login_image_url', 'https://images.unsplash.com/photo-1497366216548-37526070297c', 'URL gambar panel area masuk dan daftar. Kosongkan untuk latar gradien bawaan.', false, 'system', now())
ON CONFLICT (key) DO UPDATE
  SET value = EXCLUDED.value,
      description = EXCLUDED.description,
      updated_at = now();

-- --- Konten landing page (CMS) -------------------------------------------------
INSERT INTO landing_content (key, value, updated_by, updated_at)
VALUES
  ('hero_title', 'Temukan properti yang tepat untukmu.', 'system', now()),
  ('hero_subtitle', 'Platform listing properti terpercaya untuk sewa dan jual.', 'system', now()),
  ('hero_cta_text', 'Sewa & Beli Properti', 'system', now()),
  ('hero_cta_href', '/katalog', 'system', now()),
  ('hero_image_url', '', 'system', now()),
  ('cta_title', 'Siap memasarkan properti melalui ENM?', 'system', now()),
  ('cta_subtitle', 'Daftar sebagai mitra dan jangkau lebih banyak calon pembeli.', 'system', now()),
  ('cta_button_text', 'Daftar Sebagai Mitra', 'system', now()),
  ('cta_button_href', '/daftar-mitra', 'system', now()),
  ('meta_title', 'ENM Property - Sewa & Jual Properti Terbaik di Bali', 'system', now()),
  ('meta_description', 'Temukan properti impian untuk disewa atau dibeli. Daftar mitra untuk memasarkan properti Anda.', 'system', now()),
  ('og_image_url', '', 'system', now())
ON CONFLICT (key) DO UPDATE
  SET value = EXCLUDED.value,
      updated_at = now();

-- --- Selesai -------------------------------------------------------------------
SELECT 'Seed selesai. Baris per tabel:' AS info;
SELECT 'agreements' AS tabel, count(*) FROM agreements
UNION ALL SELECT 'plans', count(*) FROM plans
UNION ALL SELECT 'categories', count(*) FROM categories
UNION ALL SELECT 'facilities', count(*) FROM facilities
UNION ALL SELECT 'settings', count(*) FROM settings
UNION ALL SELECT 'landing_content', count(*) FROM landing_content;
