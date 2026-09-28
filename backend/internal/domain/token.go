package domain

import "time"

// AccessToken is deliberately NOT a struct that gets persisted whole — it's
// a signed JWT, so its "storage" is the JWT itself plus the signing key.
// This type exists to describe what the TokenIssuer port returns.
type AccessToken struct {
	Value     string // the signed JWT
	ExpiresAt time.Time
}

// RefreshToken, unlike AccessToken, IS persisted server-side — that's what
// makes it revocable, unlike a stateless JWT access token.
type RefreshToken struct {
	ID         string
	UserID     string
	ClientID   string
	TokenHash  string // the raw value is only ever shown once, to the client
	Scopes     []string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy *string // set on rotation; a non-nil value on an otherwise
	// "valid" token being presented again is a signal of possible theft
}

func (r RefreshToken) IsRevoked() bool { return r.RevokedAt != nil }
