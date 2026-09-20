CREATE TABLE sessions (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id),
    expires_at  TIMESTAMPTZ NOT NULL
);
