-- Tabel kategori properti untuk katalog publik dan master data superadmin.
CREATE TABLE IF NOT EXISTS categories (
    id          text        PRIMARY KEY,
    slug        text        NOT NULL,
    name        text        NOT NULL DEFAULT '',
    description text        NOT NULL DEFAULT '',
    sort_order  integer     NOT NULL DEFAULT 0,
    is_active   boolean     NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Slug kategori unik dan kolom filter umum.
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_slug ON categories (slug);
CREATE INDEX IF NOT EXISTS idx_categories_is_active ON categories (is_active);
CREATE INDEX IF NOT EXISTS idx_categories_sort_order ON categories (sort_order);
