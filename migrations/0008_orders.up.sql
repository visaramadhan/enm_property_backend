-- Tabel transaksi pembelian paket langganan (order).
CREATE TABLE IF NOT EXISTS orders (
    id                      text        PRIMARY KEY,
    code                    text        NOT NULL,
    user_id                 text        NOT NULL DEFAULT '',
    partner_id              text        NOT NULL DEFAULT '',
    partner_name            text        NOT NULL DEFAULT '',
    plan_id                 text        NOT NULL DEFAULT '',
    plan_name               text        NOT NULL DEFAULT '',
    plan_price              bigint      NOT NULL DEFAULT 0,
    plan_duration_days      integer     NOT NULL DEFAULT 0,
    plan_listing_quota      integer     NOT NULL DEFAULT 0,
    amount                  bigint      NOT NULL DEFAULT 0,
    kind                    text        NOT NULL DEFAULT '',
    status                  text        NOT NULL DEFAULT '',
    provider                text        NOT NULL DEFAULT '',
    snap_token              text        NOT NULL DEFAULT '',
    snap_redirect_url       text        NOT NULL DEFAULT '',
    midtrans_transaction_id text        NOT NULL DEFAULT '',
    midtrans_payment_type   text        NOT NULL DEFAULT '',
    paid_at                 timestamptz,
    expires_at              timestamptz NOT NULL DEFAULT now(),
    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now()
);

-- Kode order unik karena dipakai pada URL status dan invoice.
CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_code ON orders (code);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);
CREATE INDEX IF NOT EXISTS idx_orders_partner_id ON orders (partner_id);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);
CREATE INDEX IF NOT EXISTS idx_orders_plan_id ON orders (plan_id);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders (created_at);
