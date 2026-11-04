-- Tabel pengaturan sistem berbentuk pasangan kunci-nilai.
CREATE TABLE IF NOT EXISTS settings (
    key         text        PRIMARY KEY,
    value       text        NOT NULL DEFAULT '',
    description text        NOT NULL DEFAULT '',
    is_secret   boolean     NOT NULL DEFAULT false,
    updated_by  text        NOT NULL DEFAULT '',
    updated_at  timestamptz NOT NULL DEFAULT now()
);
