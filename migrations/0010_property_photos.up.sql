-- Tabel foto properti yang disimpan pada storage.
CREATE TABLE IF NOT EXISTS property_photos (
    id           text        PRIMARY KEY,
    property_id  text        NOT NULL DEFAULT '',
    url          text        NOT NULL DEFAULT '',
    storage_path text        NOT NULL DEFAULT '',
    mime_type    text        NOT NULL DEFAULT '',
    size_bytes   bigint      NOT NULL DEFAULT 0,
    width        integer     NOT NULL DEFAULT 0,
    height       integer     NOT NULL DEFAULT 0,
    sort_order   integer     NOT NULL DEFAULT 0,
    is_primary   boolean     NOT NULL DEFAULT false,
    created_at   timestamptz NOT NULL DEFAULT now()
);

-- Index untuk memuat foto berdasarkan properti dan urutan tampil.
CREATE INDEX IF NOT EXISTS idx_property_photos_property_id ON property_photos (property_id);
CREATE INDEX IF NOT EXISTS idx_property_photos_sort_order ON property_photos (property_id, sort_order);
