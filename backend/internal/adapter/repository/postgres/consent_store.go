package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type ConsentStore struct {
	pool *pgxpool.Pool
}

func NewConsentStore(pool *pgxpool.Pool) *ConsentStore {
	return &ConsentStore{pool: pool}
}

func (s *ConsentStore) Find(ctx context.Context, userID, clientID string) (domain.Consent, error) {
	row := s.pool.QueryRow(ctx, `SELECT user_id, client_id, scopes, granted_at FROM consents WHERE user_id = $1 AND client_id = $2`, userID, clientID)
	var c domain.Consent
	if err := row.Scan(&c.UserID, &c.ClientID, &c.Scopes, &c.GrantedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Consent{}, domain.NotFoundError{Message: "no consent recorded"}
		}
		return domain.Consent{}, err
	}
	return c, nil
}

// Save upserts, since a user re-approving (perhaps with additional scopes)
// should replace the prior consent record, not create a duplicate — the
// (user_id, client_id) primary key from the schema is what makes this safe.
func (s *ConsentStore) Save(ctx context.Context, c domain.Consent) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO consents (user_id, client_id, scopes, granted_at) VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, client_id) DO UPDATE SET scopes = EXCLUDED.scopes, granted_at = EXCLUDED.granted_at`,
		c.UserID, c.ClientID, c.Scopes, time.Now())
	return err
}
