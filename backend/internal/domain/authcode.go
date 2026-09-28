package domain

import "time"

// AuthorizationCode is short-lived and single-use. ConsumedAt is the field
// that enforces single-use: it starts nil, and redemption is required to be
// an atomic "set ConsumedAt if it's currently nil" operation in whatever
// store implements AuthCodeStore — never a separate read-then-delete, or
// two concurrent redemptions could both succeed.
type AuthorizationCode struct {
	Code                string
	UserID              string
	ClientID            string
	RedirectURI         string // the exact URI presented at /authorize
	Scopes              []string
	CodeChallenge       string
	CodeChallengeMethod string // "S256" — plain should be rejected in v1
	ExpiresAt           time.Time
	ConsumedAt          *time.Time
}

func (a AuthorizationCode) IsExpired(now time.Time) bool {
	return now.After(a.ExpiresAt)
}

func (a AuthorizationCode) IsConsumed() bool {
	return a.ConsumedAt != nil
}
