-- Tabel versi perjanjian kemitraan yang harus disetujui mitra.
CREATE TABLE IF NOT EXISTS agreements (
    id           text        PRIMARY KEY,
    version      text        NOT NULL DEFAULT '',
    title        text        NOT NULL DEFAULT '',
    content      text        NOT NULL DEFAULT '',
    is_active    boolean     NOT NULL DEFAULT false,
    effective_at timestamptz NOT NULL DEFAULT now(),
    created_by   text        NOT NULL DEFAULT '',
    created_at   timestamptz NOT NULL DEFAULT now()
);

-- Index untuk pencarian versi perjanjian aktif terbaru.
CREATE INDEX IF NOT EXISTS idx_agreements_is_active ON agreements (is_active);
CREATE INDEX IF NOT EXISTS idx_agreements_effective_at ON agreements (effective_at);
