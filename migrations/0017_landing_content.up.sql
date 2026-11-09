-- Tabel konten landing page (CMS) berbentuk pasangan kunci-nilai.
CREATE TABLE IF NOT EXISTS landing_content (
    key        text        PRIMARY KEY,
    value      text        NOT NULL DEFAULT '',
    updated_by text        NOT NULL DEFAULT '',
    updated_at timestamptz NOT NULL DEFAULT now()
);
