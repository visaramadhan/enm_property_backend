-- Tabel properti: listing milik mitra beserta statistik boosting dan reputasi.
CREATE TABLE IF NOT EXISTS properties (
    id               text             PRIMARY KEY,
    partner_id       text             NOT NULL DEFAULT '',
    partner_name     text             NOT NULL DEFAULT '',
    slug             text             NOT NULL,
    title            text             NOT NULL DEFAULT '',
    description      text             NOT NULL DEFAULT '',
    listing_type     text             NOT NULL DEFAULT '',
    property_type    text             NOT NULL DEFAULT '',
    category_id      text             NOT NULL DEFAULT '',
    address          text             NOT NULL DEFAULT '',
    city             text             NOT NULL DEFAULT '',
    province         text             NOT NULL DEFAULT '',
    district         text             NOT NULL DEFAULT '',
    latitude         double precision NOT NULL DEFAULT 0,
    longitude        double precision NOT NULL DEFAULT 0,
    price            bigint           NOT NULL DEFAULT 0,
    price_period     text             NOT NULL DEFAULT '',
    land_area        integer          NOT NULL DEFAULT 0,
    building_area    integer          NOT NULL DEFAULT 0,
    bedrooms         integer          NOT NULL DEFAULT 0,
    bathrooms        integer          NOT NULL DEFAULT 0,
    facilities       jsonb            NOT NULL DEFAULT '[]'::jsonb,
    units_total      integer          NOT NULL DEFAULT 0,
    units_available  integer          NOT NULL DEFAULT 0,
    status           text             NOT NULL DEFAULT '',
    views            integer          NOT NULL DEFAULT 0,
    contact_count    integer          NOT NULL DEFAULT 0,
    boost_active     boolean          NOT NULL DEFAULT false,
    boost_until      timestamptz,
    rating_count     integer          NOT NULL DEFAULT 0,
    rating_avg       numeric(3,2)     NOT NULL DEFAULT 0,
    favorite_count   integer          NOT NULL DEFAULT 0,
    published_at     timestamptz,
    deleted_at       timestamptz,
    created_at       timestamptz      NOT NULL DEFAULT now(),
    updated_at       timestamptz      NOT NULL DEFAULT now()
);

-- Slug unik dan index kolom yang sering difilter katalog maupun laporan.
CREATE UNIQUE INDEX IF NOT EXISTS idx_properties_slug ON properties (slug);
CREATE INDEX IF NOT EXISTS idx_properties_partner_id ON properties (partner_id);
CREATE INDEX IF NOT EXISTS idx_properties_status ON properties (status);
CREATE INDEX IF NOT EXISTS idx_properties_category_id ON properties (category_id);
CREATE INDEX IF NOT EXISTS idx_properties_city ON properties (lower(city));
CREATE INDEX IF NOT EXISTS idx_properties_province ON properties (lower(province));
CREATE INDEX IF NOT EXISTS idx_properties_listing_type ON properties (listing_type);
CREATE INDEX IF NOT EXISTS idx_properties_property_type ON properties (property_type);
CREATE INDEX IF NOT EXISTS idx_properties_boost_active ON properties (boost_active);
CREATE INDEX IF NOT EXISTS idx_properties_deleted_at ON properties (deleted_at);
CREATE INDEX IF NOT EXISTS idx_properties_created_at ON properties (created_at);
