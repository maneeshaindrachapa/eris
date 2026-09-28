-- Dev/test seed data only — never run this against a real environment.
-- Passwords: alice/correct-horse, bob/battery-staple.

INSERT INTO users (id, email, password_hash, email_verified, status) VALUES
    ('user-1', 'alice@example.com', '$2a$10$5ivjsqa0ELNCXp0yif/0hec/ambZDxYttJKNb/rRfamTpBOJtLn/m', true, 'active'),
    ('user-2', 'bob@example.com',   '$2a$10$1EpGuYgJ56afMz3CwATpauBqDLJXLiWkHPnRUBnjHsr4i9F/YzvOq', true, 'active');

INSERT INTO clients (id, name, type, client_secret_hash, redirect_uris, allowed_scopes, allowed_grant_types) VALUES
    ('spa-client', 'Demo SPA', 'public', NULL,
     ARRAY['http://localhost:3000/callback'],
     ARRAY['openid', 'profile', 'email'],
     ARRAY['authorization_code', 'refresh_token']);
