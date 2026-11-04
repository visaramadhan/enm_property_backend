-- Tabel bukti persetujuan perjanjian oleh mitra.
CREATE TABLE IF NOT EXISTS agreement_acceptances (
    id                text        PRIMARY KEY,
    agreement_id      text        NOT NULL DEFAULT '',
    agreement_version text        NOT NULL DEFAULT '',
    user_id           text        NOT NULL DEFAULT '',
    partner_id        text        NOT NULL DEFAULT '',
    accepted_at       timestamptz NOT NULL DEFAULT now(),
    ip_address        text        NOT NULL DEFAULT '',
    user_agent        text        NOT NULL DEFAULT ''
);

-- Index untuk menelusuri persetujuan per mitra dan per versi perjanjian.
CREATE INDEX IF NOT EXISTS idx_agreement_acceptances_partner_id ON agreement_acceptances (partner_id);
CREATE INDEX IF NOT EXISTS idx_agreement_acceptances_agreement_id ON agreement_acceptances (agreement_id);
CREATE INDEX IF NOT EXISTS idx_agreement_acceptances_user_id ON agreement_acceptances (user_id);
CREATE INDEX IF NOT EXISTS idx_agreement_acceptances_accepted_at ON agreement_acceptances (accepted_at);
