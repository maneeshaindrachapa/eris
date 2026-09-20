package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type SessionStore struct {
	pool *pgxpool.Pool
}

func NewSessionStore(pool *pgxpool.Pool) *SessionStore {
	return &SessionStore{pool: pool}
}

func (s *SessionStore) Save(ctx context.Context, sess domain.Session) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET expires_at = EXCLUDED.expires_at`,
		sess.ID, sess.UserID, sess.ExpiresAt)
	return err
}

func (s *SessionStore) Find(ctx context.Context, id string) (domain.Session, error) {
	row := s.pool.QueryRow(ctx, `SELECT id, user_id, expires_at FROM sessions WHERE id = $1`, id)
	var sess domain.Session
	if err := row.Scan(&sess.ID, &sess.UserID, &sess.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Session{}, domain.NotFoundError{Message: "session not found"}
		}
		return domain.Session{}, err
	}
	return sess, nil
}
