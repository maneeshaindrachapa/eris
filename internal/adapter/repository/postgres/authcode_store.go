package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type AuthCodeStore struct {
	pool *pgxpool.Pool
}

func NewAuthCodeStore(pool *pgxpool.Pool) *AuthCodeStore {
	return &AuthCodeStore{pool: pool}
}

func (s *AuthCodeStore) Save(ctx context.Context, ac domain.AuthorizationCode) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO authorization_codes (code, user_id, client_id, redirect_uri, scopes, code_challenge, code_challenge_method, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		ac.Code, ac.UserID, ac.ClientID, ac.RedirectURI, ac.Scopes, ac.CodeChallenge, ac.CodeChallengeMethod, ac.ExpiresAt)
	return err
}

// Consume is ONE round trip: an UPDATE that only succeeds if consumed_at is
// still NULL, returning the row atomically. This is the SQL equivalent of
// the mutex-guarded map write in the in-memory version — Postgres's row
// lock during the UPDATE is what prevents two concurrent redemptions from
// both succeeding, without needing an app-level lock at all.
func (s *AuthCodeStore) Consume(ctx context.Context, code string, now time.Time) (domain.AuthorizationCode, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE authorization_codes
		SET consumed_at = $2
		WHERE code = $1 AND consumed_at IS NULL
		RETURNING code, user_id, client_id, redirect_uri, scopes, code_challenge, code_challenge_method, expires_at, consumed_at`,
		code, now)

	var ac domain.AuthorizationCode
	if err := row.Scan(&ac.Code, &ac.UserID, &ac.ClientID, &ac.RedirectURI, &ac.Scopes, &ac.CodeChallenge, &ac.CodeChallengeMethod, &ac.ExpiresAt, &ac.ConsumedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Zero rows updated means EITHER the code doesn't exist OR it
			// was already consumed — we deliberately can't tell which from
			// here, and that's fine: both cases return the same
			// invalid_grant to the caller, so there's no information leak
			// about whether a guessed code ever existed.
			return domain.AuthorizationCode{}, domain.InvalidGrantError{Message: "authorization code not found or already used"}
		}
		return domain.AuthorizationCode{}, err
	}
	if now.After(ac.ExpiresAt) {
		return domain.AuthorizationCode{}, domain.InvalidGrantError{Message: "authorization code expired"}
	}
	return ac, nil
}
