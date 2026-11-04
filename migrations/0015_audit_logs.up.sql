-- Tabel catatan audit aksi admin dan superadmin.
CREATE TABLE IF NOT EXISTS audit_logs (
    id          text        PRIMARY KEY,
    actor_id    text        NOT NULL DEFAULT '',
    actor_name  text        NOT NULL DEFAULT '',
    actor_email text        NOT NULL DEFAULT '',
    actor_role  text        NOT NULL DEFAULT '',
    action      text        NOT NULL DEFAULT '',
    entity      text        NOT NULL DEFAULT '',
    entity_id   text        NOT NULL DEFAULT '',
    "before"    text        NOT NULL DEFAULT '',
    "after"     text        NOT NULL DEFAULT '',
    note        text        NOT NULL DEFAULT '',
    ip_address  text        NOT NULL DEFAULT '',
    user_agent  text        NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- Index untuk menelusuri audit per entitas, aktor, dan waktu.
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs (lower(entity));
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_id ON audit_logs (entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs (actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_role ON audit_logs (actor_role);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs (action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at);
