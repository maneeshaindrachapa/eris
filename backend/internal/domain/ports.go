package domain

import (
	"context"
	"time"
)

// ---- Driven ports ----
// The application layer calls these; the memory and Postgres adapters both
// implement the same interfaces, so higher layers do not change when the
// configured storage backend changes.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
}

type ClientRepository interface {
	FindByID(ctx context.Context, id string) (Client, error)
	List(ctx context.Context) ([]Client, error)
	Create(ctx context.Context, client Client) error
	Update(ctx context.Context, client Client) error
}

// AuthCodeStore's Consume method is deliberately one atomic operation, not
// Find-then-Delete, to make single-use enforcement safe under concurrency —
// see the note in authcode.go.
type AuthCodeStore interface {
	Save(ctx context.Context, code AuthorizationCode) error
	Consume(ctx context.Context, code string, now time.Time) (AuthorizationCode, error)
}

type RefreshTokenStore interface {
	Save(ctx context.Context, rt RefreshToken) error
	FindByTokenHash(ctx context.Context, hash string) (RefreshToken, error)
	Revoke(ctx context.Context, id string, now time.Time) error
}

type SessionStore interface {
	Save(ctx context.Context, s Session) error
	Find(ctx context.Context, id string) (Session, error)
}

type ConsentStore interface {
	Find(ctx context.Context, userID, clientID string) (Consent, error)
	Save(ctx context.Context, c Consent) error
}

// TokenIssuer hides all JWT/crypto mechanics from the application layer —
// it just asks for a token "for this user, this client, these scopes."
type TokenIssuer interface {
	IssueAccessToken(ctx context.Context, userID, clientID string, scopes []string) (AccessToken, error)
	IssueIDToken(ctx context.Context, userID, clientID string, nonce string) (string, error)
}

// ---- Driving ports ----
// What the HTTP adapter (and later, any other inbound adapter) calls INTO.

type AuthorizeRequest struct {
	ClientID            string
	RedirectURI         string
	Scopes              []string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	SessionID           string // empty if the browser has no Eris session yet
}

type AuthorizeService interface {
	// Authorize validates the request and, if a valid session+consent exist,
	// returns an issued code. Otherwise it returns a status telling the
	// HTTP adapter whether to render a login form or a consent screen.
	Authorize(ctx context.Context, req AuthorizeRequest) (AuthorizeResult, error)
	Login(ctx context.Context, email, password string) (sessionID string, err error)
	ApproveConsent(ctx context.Context, req AuthorizeRequest, approvedScopes []string) (code string, err error)
}

type AuthorizeResult struct {
	NeedsLogin   bool
	NeedsConsent bool
	Code         string // set only when neither of the above is true
}

type TokenService interface {
	ExchangeAuthorizationCode(ctx context.Context, code, redirectURI, codeVerifier, clientID, clientSecret string) (TokenResult, error)
	RefreshAccessToken(ctx context.Context, refreshToken, clientID, clientSecret string) (TokenResult, error)
}

type TokenResult struct {
	AccessToken  string
	RefreshToken string
	IDToken      string // empty unless "openid" scope was requested
	ExpiresIn    int
	Scopes       []string
}

type AdminClientInput struct {
	Name              string
	Type              ClientType
	RedirectURIs      []string
	AllowedScopes     []string
	AllowedGrantTypes []GrantType
}

type AdminClientResult struct {
	Client       Client
	ClientSecret string
}

type AdminService interface {
	ListClients(ctx context.Context) ([]Client, error)
	CreateClient(ctx context.Context, input AdminClientInput) (AdminClientResult, error)
	UpdateClient(ctx context.Context, id string, input AdminClientInput) (Client, error)
}

type KeyStore interface {
	// ActiveKey returns the key currently used to sign NEW tokens.
	// If none exists yet, the caller is expected to generate and save one.
	ActiveKey(ctx context.Context) (kid string, privateKeyPEM string, err error)
	Save(ctx context.Context, kid, algorithm, privateKeyPEM, publicKeyPEM string) error
	// AllPublicKeys returns every non-expired key (active + retired-but-
	// still-verifiable) — this is exactly what the JWKS endpoint will serve.
	AllPublicKeys(ctx context.Context) (map[string]string, error) // kid -> public PEM
}
