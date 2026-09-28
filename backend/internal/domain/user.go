package domain

import "time"

// User is the resource owner. It holds nothing about sessions, tokens, or
// clients — just identity. PasswordHash is exactly that: a hash (bcrypt or
// argon2), never a plaintext password, not even transiently held longer
// than the login check requires.
type User struct {
	ID            string
	Email         string
	PasswordHash  string
	EmailVerified bool
	Status        UserStatus
	CreatedAt     time.Time
}

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
	UserStatusLocked UserStatus = "locked"
)
