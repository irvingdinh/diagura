-- +migration Up
CREATE TABLE roles (
    id         TEXT PRIMARY KEY,
    slug       TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    role_id       TEXT NOT NULL REFERENCES roles(id),
    email         TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL,
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL,
    deleted_at    TEXT
);

CREATE TABLE sessions (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id),
    token_hash          TEXT NOT NULL UNIQUE,
    expires_at          TEXT NOT NULL,
    absolute_expires_at TEXT NOT NULL,
    ip_address          TEXT NOT NULL DEFAULT '',
    user_agent          TEXT NOT NULL DEFAULT '',
    created_at          TEXT NOT NULL,
    updated_at          TEXT NOT NULL
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);

INSERT INTO roles (id, slug, name, created_at, updated_at)
VALUES ('00000000-0000-7000-0000-000000000000', 'super_admin', 'Super Admin', '2026-01-01 00:00:00.000', '2026-01-01 00:00:00.000');

INSERT INTO roles (id, slug, name, created_at, updated_at)
VALUES ('00000000-0000-7000-0000-000000000001', 'admin', 'Admin', '2026-01-01 00:00:00.000', '2026-01-01 00:00:00.000');

INSERT INTO roles (id, slug, name, created_at, updated_at)
VALUES ('00000000-0000-7000-0000-000000000002', 'user', 'User', '2026-01-01 00:00:00.000', '2026-01-01 00:00:00.000');

INSERT INTO users (id, role_id, email, name, password_hash, created_at, updated_at)
VALUES ('00000000-0000-7000-0001-000000000000', '00000000-0000-7000-0000-000000000000', 'john.doe@example.com', 'John Doe', '$argon2id$v=19$m=19456,t=2,p=1$cFRG9DskwaRlt8sac0c5qQ$PT/6xa4fNr5ZvlgNlx3AU0deOroX+AEwIRSleLJQ/TM', '2026-01-01 00:00:00.000', '2026-01-01 00:00:00.000');

-- +migration Down
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
