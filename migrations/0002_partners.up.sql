-- Tabel mitra: profil kemitraan yang dimiliki pengguna berperan mitra.
CREATE TABLE IF NOT EXISTS partners (
    id                    text        PRIMARY KEY,
    user_id               text        NOT NULL DEFAULT '',
    business_name         text        NOT NULL DEFAULT '',
    pic_name              text        NOT NULL DEFAULT '',
    phone                 text        NOT NULL DEFAULT '',
    email                 text        NOT NULL DEFAULT '',
    address               text        NOT NULL DEFAULT '',
    city                  text        NOT NULL DEFAULT '',
    province              text        NOT NULL DEFAULT '',
    status                text        NOT NULL DEFAULT '',
    suspended_reason      text        NOT NULL DEFAULT '',
    agreement_accepted_at timestamptz,
    joined_at             timestamptz,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

-- Index untuk penautan ke pengguna dan filter laporan.
CREATE INDEX IF NOT EXISTS idx_partners_user_id ON partners (user_id);
CREATE INDEX IF NOT EXISTS idx_partners_status ON partners (status);
CREATE INDEX IF NOT EXISTS idx_partners_city ON partners (lower(city));
CREATE INDEX IF NOT EXISTS idx_partners_created_at ON partners (created_at);
