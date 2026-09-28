CREATE TABLE refresh_tokens (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id),
    client_id    TEXT NOT NULL REFERENCES clients(id),
    token_hash   TEXT NOT NULL UNIQUE,
    scopes       TEXT[] NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ,
    replaced_by  TEXT
);
