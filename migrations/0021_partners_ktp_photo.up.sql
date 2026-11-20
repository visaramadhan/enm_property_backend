-- Menambah kolom foto KTP mitra yang wajib diunggah saat pendaftaran mitra.
ALTER TABLE partners
    ADD COLUMN IF NOT EXISTS ktp_photo_url text NOT NULL DEFAULT '';
