-- Master data fasilitas properti (mis. Carport, Kolam Renang, Keamanan 24 jam)
-- yang dipilih mitra saat mengunggah listing dan dikelola superadmin.
CREATE TABLE IF NOT EXISTS facilities (
    id          text        PRIMARY KEY,
    slug        text        NOT NULL,
    name        text        NOT NULL DEFAULT '',
    description text        NOT NULL DEFAULT '',
    icon        text        NOT NULL DEFAULT '',
    sort_order  integer     NOT NULL DEFAULT 0,
    is_active   boolean     NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_facilities_slug ON facilities (slug);
CREATE INDEX IF NOT EXISTS idx_facilities_is_active ON facilities (is_active);
CREATE INDEX IF NOT EXISTS idx_facilities_sort_order ON facilities (sort_order);
