-- Tabel pengguna: akun yang disinkronkan dari Firebase Authentication.
CREATE TABLE IF NOT EXISTS users (
    id             text        PRIMARY KEY,
    firebase_uid   text        NOT NULL DEFAULT '',
    email          text        NOT NULL DEFAULT '',
    name           text        NOT NULL DEFAULT '',
    phone          text        NOT NULL DEFAULT '',
    avatar_url     text        NOT NULL DEFAULT '',
    role           text        NOT NULL DEFAULT '',
    status         text        NOT NULL DEFAULT '',
    email_verified boolean     NOT NULL DEFAULT false,
    last_login_at  timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- Index untuk pencarian akun berdasarkan identitas Firebase, email, dan filter umum.
CREATE INDEX IF NOT EXISTS idx_users_firebase_uid ON users (firebase_uid);
CREATE INDEX IF NOT EXISTS idx_users_email ON users (lower(email));
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users (created_at);
