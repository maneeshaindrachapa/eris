package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type ClientRepository struct {
	pool *pgxpool.Pool
}

func NewClientRepository(pool *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{pool: pool}
}

func (r *ClientRepository) FindByID(ctx context.Context, id string) (domain.Client, error) {
	return r.ScanOne(ctx, `SELECT id, name, type, COALESCE(client_secret_hash, ''), redirect_uris, allowed_scopes, allowed_grant_types
		FROM clients WHERE id = $1`, id)
}

func (r *ClientRepository) ScanOne(ctx context.Context, query string, args string) (domain.Client, error) {
	var c domain.Client
	var clientType string
	var grantTypes []string
	row := r.pool.QueryRow(ctx, query, args)
	if err := row.Scan(&c.ID, &c.Name, &clientType, &c.ClientSecretHash, &c.RedirectURIs, &c.AllowedScopes, &grantTypes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Client{}, domain.NotFoundError{Message: "client not found"}
		}
		return domain.Client{}, err
	}
	c.Type = domain.ClientType(clientType)
	for _, g := range grantTypes {
		c.AllowedGrantTypes = append(c.AllowedGrantTypes, domain.GrantType(g))
	}
	return c, nil
}
