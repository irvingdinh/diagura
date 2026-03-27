-- +migration Up
CREATE TABLE users (
    id         TEXT PRIMARY KEY,
    email      TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- +migration Down
DROP TABLE IF EXISTS users;
