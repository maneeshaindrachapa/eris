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

func (r *ClientRepository) List(ctx context.Context) ([]domain.Client, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, type, COALESCE(client_secret_hash, ''), redirect_uris, allowed_scopes, allowed_grant_types FROM clients ORDER BY name, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	clients := make([]domain.Client, 0)
	for rows.Next() {
		client, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	return clients, rows.Err()
}

func (r *ClientRepository) Create(ctx context.Context, client domain.Client) error {
	grants := make([]string, len(client.AllowedGrantTypes))
	for i, grant := range client.AllowedGrantTypes {
		grants[i] = string(grant)
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO clients (id, name, type, client_secret_hash, redirect_uris, allowed_scopes, allowed_grant_types) VALUES ($1,$2,$3,NULLIF($4,''),$5,$6,$7)`, client.ID, client.Name, client.Type, client.ClientSecretHash, client.RedirectURIs, client.AllowedScopes, grants)
	return err
}

func (r *ClientRepository) Update(ctx context.Context, client domain.Client) error {
	grants := make([]string, len(client.AllowedGrantTypes))
	for i, grant := range client.AllowedGrantTypes {
		grants[i] = string(grant)
	}
	command, err := r.pool.Exec(ctx, `UPDATE clients SET name=$2, redirect_uris=$3, allowed_scopes=$4, allowed_grant_types=$5 WHERE id=$1`, client.ID, client.Name, client.RedirectURIs, client.AllowedScopes, grants)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return domain.ErrorNotFound("client not found")
	}
	return nil
}

func (r *ClientRepository) ScanOne(ctx context.Context, query string, args string) (domain.Client, error) {
	row := r.pool.QueryRow(ctx, query, args)
	c, err := scanClient(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Client{}, domain.NotFoundError{Message: "client not found"}
		}
		return domain.Client{}, err
	}
	return c, nil
}

type clientScanner interface{ Scan(dest ...any) error }

func scanClient(row clientScanner) (domain.Client, error) {
	var c domain.Client
	var clientType string
	var grantTypes []string
	if err := row.Scan(&c.ID, &c.Name, &clientType, &c.ClientSecretHash, &c.RedirectURIs, &c.AllowedScopes, &grantTypes); err != nil {
		return domain.Client{}, err
	}
	c.Type = domain.ClientType(clientType)
	for _, g := range grantTypes {
		c.AllowedGrantTypes = append(c.AllowedGrantTypes, domain.GrantType(g))
	}
	return c, nil
}
