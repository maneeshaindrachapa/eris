CREATE TABLE authorization_codes (
    code                   TEXT PRIMARY KEY,
    user_id                TEXT NOT NULL REFERENCES users(id),
    client_id              TEXT NOT NULL REFERENCES clients(id),
    redirect_uri           TEXT NOT NULL,
    scopes                 TEXT[] NOT NULL,
    code_challenge         TEXT NOT NULL,
    code_challenge_method  TEXT NOT NULL,
    expires_at             TIMESTAMPTZ NOT NULL,
    consumed_at            TIMESTAMPTZ           -- NULL until redeemed; the single-use flag
);
