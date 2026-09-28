package domain

import "time"

// Session represents the user's authenticated browser session with Eris
// ITSELF — separate from any client's tokens. This is what lets a second
// /authorize call for a different client skip the login form.
type Session struct {
	ID        string // the cookie value
	UserID    string
	ExpiresAt time.Time
}

func (s Session) IsExpired(now time.Time) bool { return now.After(s.ExpiresAt) }
