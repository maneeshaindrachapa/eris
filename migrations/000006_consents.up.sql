CREATE TABLE consents (
    user_id     TEXT NOT NULL REFERENCES users(id),
    client_id   TEXT NOT NULL REFERENCES clients(id),
    scopes      TEXT[] NOT NULL,
    granted_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, client_id)
);
