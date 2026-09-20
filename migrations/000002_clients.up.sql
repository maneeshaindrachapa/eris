CREATE TABLE clients (
    id                  TEXT PRIMARY KEY,
    name                TEXT NOT NULL,
    type                TEXT NOT NULL,           -- 'confidential' | 'public'
    client_secret_hash  TEXT,                    -- null for public clients
    redirect_uris       TEXT[] NOT NULL,
    allowed_scopes      TEXT[] NOT NULL,
    allowed_grant_types TEXT[] NOT NULL
);
