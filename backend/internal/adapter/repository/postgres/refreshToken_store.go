package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type RefreshTokenStore struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenStore(pool *pgxpool.Pool) *RefreshTokenStore {
	return &RefreshTokenStore{pool: pool}
}

func (s *RefreshTokenStore) Save(ctx context.Context, rt domain.RefreshToken) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, client_id, token_hash, scopes, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		rt.ID, rt.UserID, rt.ClientID, rt.TokenHash, rt.Scopes, rt.ExpiresAt)
	return err
}

func (s *RefreshTokenStore) FindByTokenHash(ctx context.Context, hash string) (domain.RefreshToken, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, user_id, client_id, token_hash, scopes, expires_at, revoked_at, replaced_by
		FROM refresh_tokens WHERE token_hash = $1`, hash)

	var rt domain.RefreshToken
	if err := row.Scan(&rt.ID, &rt.UserID, &rt.ClientID, &rt.TokenHash, &rt.Scopes, &rt.ExpiresAt, &rt.RevokedAt, &rt.ReplacedBy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.RefreshToken{}, domain.NotFoundError{Message: "refresh token not found"}
		}
		return domain.RefreshToken{}, err
	}
	return rt, nil
}

func (s *RefreshTokenStore) Revoke(ctx context.Context, id string, now time.Time) error {
	tag, err := s.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1 AND revoked_at IS NULL`, id, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFoundError{Message: "refresh token not found or already revoked"}
	}
	return nil
}
