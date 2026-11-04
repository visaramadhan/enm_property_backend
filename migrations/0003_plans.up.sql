-- Tabel paket langganan beserta hak akses (level, boost, keyword, katalog, WA).
CREATE TABLE IF NOT EXISTS plans (
    id             text        PRIMARY KEY,
    slug           text        NOT NULL,
    name           text        NOT NULL DEFAULT '',
    description    text        NOT NULL DEFAULT '',
    price          bigint      NOT NULL DEFAULT 0,
    duration_days  integer     NOT NULL DEFAULT 0,
    listing_quota  integer     NOT NULL DEFAULT 0,
    features       jsonb       NOT NULL DEFAULT '[]'::jsonb,
    is_free        boolean     NOT NULL DEFAULT false,
    is_active      boolean     NOT NULL DEFAULT false,
    sort_order     integer     NOT NULL DEFAULT 0,
    level          integer     NOT NULL DEFAULT 1,
    boost_enabled  boolean     NOT NULL DEFAULT false,
    keyword_access boolean     NOT NULL DEFAULT false,
    catalog_access boolean     NOT NULL DEFAULT false,
    whatsapp_access boolean    NOT NULL DEFAULT false,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- Slug paket unik agar dapat dipakai pada URL publik.
CREATE UNIQUE INDEX IF NOT EXISTS idx_plans_slug ON plans (slug);
CREATE INDEX IF NOT EXISTS idx_plans_is_active ON plans (is_active);
CREATE INDEX IF NOT EXISTS idx_plans_sort_order ON plans (sort_order);
