package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (domain.User, error) {
	return r.scanOne(ctx, `SELECT id, email, password_hash, email_verified, status, created_at FROM users WHERE id = $1`, id)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.scanOne(ctx, `SELECT id, email, password_hash, email_verified, status, created_at FROM users WHERE email = $1`, email)
}

func (r *UserRepository) scanOne(ctx context.Context, query string, arg string) (domain.User, error) {
	row := r.pool.QueryRow(ctx, query, arg)
	var u domain.User
	var status string
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.EmailVerified, &status, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.NotFoundError{Message: "user not found"}
		}
		return domain.User{}, err
	}
	u.Status = domain.UserStatus(status)
	return u, nil
}
