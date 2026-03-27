-- +migration Up
ALTER TABLE users ADD COLUMN force_password_change INTEGER NOT NULL DEFAULT 0;

-- +migration Down
ALTER TABLE users DROP COLUMN force_password_change;
