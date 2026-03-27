-- +migration Up
CREATE TABLE events (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    actor_id    TEXT,
    request_id  TEXT,
    ip          TEXT,
    entity_type TEXT,
    entity_id   TEXT,
    data        TEXT NOT NULL,
    created_at  TEXT NOT NULL
);

CREATE INDEX idx_events_name ON events(name);
CREATE INDEX idx_events_actor_id ON events(actor_id);
CREATE INDEX idx_events_entity ON events(entity_type, entity_id);
CREATE INDEX idx_events_created_at ON events(created_at);

-- +migration Down
DROP TABLE IF EXISTS events;
