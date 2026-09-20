package domain

import "time"

// Consent records that a user approved a client for a set of scopes, so
// Eris doesn't re-prompt on every login once approval already happened.
type Consent struct {
	UserID    string
	ClientID  string
	Scopes    []string
	GrantedAt time.Time
}
