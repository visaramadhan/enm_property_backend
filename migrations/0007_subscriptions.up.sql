-- Tabel langganan: masa aktif paket yang dimiliki mitra.
CREATE TABLE IF NOT EXISTS subscriptions (
    id         text        PRIMARY KEY,
    partner_id text        NOT NULL DEFAULT '',
    plan_id    text        NOT NULL DEFAULT '',
    plan_name  text        NOT NULL DEFAULT '',
    order_id   text        NOT NULL DEFAULT '',
    starts_at  timestamptz NOT NULL DEFAULT now(),
    ends_at    timestamptz NOT NULL DEFAULT now(),
    status     text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Index untuk menentukan langganan aktif per mitra.
CREATE INDEX IF NOT EXISTS idx_subscriptions_partner_id ON subscriptions (partner_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_id ON subscriptions (plan_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions (status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_ends_at ON subscriptions (ends_at);
CREATE INDEX IF NOT EXISTS idx_subscriptions_created_at ON subscriptions (created_at);
