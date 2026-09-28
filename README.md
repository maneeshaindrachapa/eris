# Eris Identity Server

Eris is a small OAuth 2.0 and OpenID Connect identity server with a PostgreSQL-backed administration portal. It implements Authorization Code with mandatory PKCE (`S256`), browser login and consent, signed JWT access and ID tokens, and rotating refresh tokens.

This project is designed for learning and local development. The admin API is not authenticated yet and must not be exposed publicly.

## Repository layout

| Path | Responsibility |
| --- | --- |
| `backend` | Go authorization server, admin API, migrations, and operational scripts |
| `frontend` | Next.js identity-server administration portal and PKCE test client |
| `config.toml` | Shared frontend, backend, PostgreSQL, and tooling configuration |

## Shared configuration

Both applications read the root `config.toml` file. The Go backend loads and validates it with [`github.com/maneeshaindrachapa/zenith`](https://github.com/maneeshaindrachapa/zenith):

```toml
environment = "development"

[backend]
storage_backend = "postgres"
server_address = ":8080"
issuer_url = "http://localhost:8080"

[frontend]
origin = "http://localhost:3000"
public_url = "http://localhost:3000"
default_callback_path = "/callback"

[postgres]
dsn = "postgres://eris:eris@localhost:5432/eris?sslmode=disable"
container = "eris-postgres"
image = "postgres:16"
user = "eris"
password = "eris"
database = "eris"
hostname = "localhost"
port = 5432
volume = "eris-postgres-data"
```

The backend looks for `config.toml` in the current or parent directory. Set `ERIS_CONFIG` only when running with a different configuration file.

## Run locally

Requirements:

- Go `1.27.1`, or a compatible toolchain
- Node.js and npm
- Docker

Start PostgreSQL and apply migrations:

```bash
./backend/scripts/connect-postgres.sh --wait
./backend/scripts/migrate.sh up
```

Start the backend:

```bash
cd backend
go run ./cmd/api
```

Start the frontend in another terminal:

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:3000`.

## Administration portal

The Applications screen reads registered OAuth clients from PostgreSQL through the backend admin API. It supports:

- Listing and searching applications
- Registering public and confidential clients
- Configuring exact redirect URIs
- Configuring allowed scopes and grant types
- Editing existing client metadata
- Showing a confidential client secret once at creation
- Launching an interactive PKCE test for public clients

Public clients do not receive a secret. Confidential-client secrets are returned once and only their bcrypt hashes are stored.

### Admin API

| Method | Endpoint | Purpose |
| --- | --- | --- |
| `GET` | `/admin/clients` | List registered clients |
| `POST` | `/admin/clients` | Register a client |
| `PUT` | `/admin/clients/:id` | Update client metadata |

The development admin API currently has no administrator authentication or authorization. Add that boundary before any non-local deployment.

## OAuth endpoints

| Method | Endpoint | Purpose |
| --- | --- | --- |
| `GET` | `/authorize` | Validate a PKCE request and show login or consent |
| `POST` | `/authorize/login` | Authenticate the resource owner |
| `POST` | `/authorize/consent` | Record consent and issue an authorization code |
| `POST` | `/token` | Exchange a code or rotate a refresh token |
| `GET` | `/healthz` | Server health check |

The PKCE test screen creates a high-entropy verifier in the browser and sends only its SHA-256 challenge through `/authorize`:

```text
code_challenge = BASE64URL_NO_PADDING(SHA256(code_verifier))
```

The verifier remains in browser session storage until the callback exchanges the one-time authorization code.

## Development data

Migrations seed one public client and two users:

| Client ID | Redirect URI | Scopes |
| --- | --- | --- |
| `spa-client` | `http://localhost:3000/callback` | `openid profile email` |

| Email | Password |
| --- | --- |
| `alice@example.com` | `correct-horse` |
| `bob@example.com` | `battery-staple` |

## Verification

Run backend tests:

```bash
cd backend
go test ./...
```

Run frontend checks:

```bash
cd frontend
npm run lint
npm run build
```

Run the complete PKCE and refresh-rotation test while the backend is running:

```bash
bash backend/scripts/eris-authcode-pkce-test.sh
```

## Database operations

```bash
./backend/scripts/migrate.sh up
./backend/scripts/migrate.sh version
./backend/scripts/migrate.sh down 1
./backend/scripts/migrate.sh goto 8
./backend/scripts/migrate.sh create add_field
./backend/scripts/connect-postgres.sh
```

The legacy client-registration CLI remains available at `backend/scripts/create-client.sh`, though the admin portal is now the normal registration workflow.

## Current limitations

- The admin API has no administrator authentication yet.
- There is no discovery document or JWKS endpoint.
- Login and consent still use minimal inline HTML forms.
- The consent form does not yet have a dedicated CSRF token.
- PKCE verifier syntax and length are not explicitly validated.
- ID-token `nonce` is not connected to the authorization request.
- Refresh revoke-and-issue is not one database transaction.
- Signing-key rotation and encryption at rest are not implemented.
