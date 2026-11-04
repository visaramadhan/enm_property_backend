-- Tabel favorit: properti yang disimpan customer.
CREATE TABLE IF NOT EXISTS favorites (
    id          text        PRIMARY KEY,
    user_id     text        NOT NULL DEFAULT '',
    property_id text        NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- Satu pasangan pengguna-properti hanya boleh disimpan sekali.
CREATE UNIQUE INDEX IF NOT EXISTS idx_favorites_user_property ON favorites (user_id, property_id);
CREATE INDEX IF NOT EXISTS idx_favorites_property_id ON favorites (property_id);
CREATE INDEX IF NOT EXISTS idx_favorites_created_at ON favorites (created_at);
