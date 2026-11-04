-- Tabel tayangan halaman detail properti untuk statistik mitra.
CREATE TABLE IF NOT EXISTS property_views (
    id          text        PRIMARY KEY,
    property_id text        NOT NULL DEFAULT '',
    partner_id  text        NOT NULL DEFAULT '',
    viewed_at   timestamptz NOT NULL DEFAULT now()
);

-- Index untuk rekap tayangan per properti/mitra.
CREATE INDEX IF NOT EXISTS idx_property_views_property_id ON property_views (property_id);
CREATE INDEX IF NOT EXISTS idx_property_views_partner_id ON property_views (partner_id);
CREATE INDEX IF NOT EXISTS idx_property_views_viewed_at ON property_views (viewed_at);
