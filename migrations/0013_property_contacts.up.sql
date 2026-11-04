-- Tabel lead WhatsApp yang diteruskan ke mitra.
CREATE TABLE IF NOT EXISTS property_contacts (
    id          text        PRIMARY KEY,
    property_id text        NOT NULL DEFAULT '',
    partner_id  text        NOT NULL DEFAULT '',
    user_id     text        NOT NULL DEFAULT '',
    channel     text        NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- Index untuk rekap lead per properti/mitra.
CREATE INDEX IF NOT EXISTS idx_property_contacts_property_id ON property_contacts (property_id);
CREATE INDEX IF NOT EXISTS idx_property_contacts_partner_id ON property_contacts (partner_id);
CREATE INDEX IF NOT EXISTS idx_property_contacts_created_at ON property_contacts (created_at);
