CREATE TABLE signing_keys (
    kid          TEXT PRIMARY KEY,
    algorithm    TEXT NOT NULL,
    private_key  TEXT NOT NULL,   -- PEM-encoded; encrypt at rest in prod (see note below)
    public_key   TEXT NOT NULL,   -- PEM-encoded
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    retired_at   TIMESTAMPTZ      -- NULL = still used for signing new tokens
);
