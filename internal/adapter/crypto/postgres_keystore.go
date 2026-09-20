package crypto

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type PostgresKeyStore struct {
	pool *pgxpool.Pool
}

func NewPostgresKeyStore(pool *pgxpool.Pool) *PostgresKeyStore {
	return &PostgresKeyStore{pool: pool}
}

func (k *PostgresKeyStore) ActiveKey(ctx context.Context) (string, string, error) {
	row := k.pool.QueryRow(ctx, `SELECT kid, private_key FROM signing_keys WHERE retired_at IS NULL ORDER BY created_at DESC LIMIT 1`)
	var kid, priv string
	if err := row.Scan(&kid, &priv); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", domain.NotFoundError{Message: "no active signing key"}
		}
		return "", "", err
	}
	return kid, priv, nil
}

func (k *PostgresKeyStore) Save(ctx context.Context, kid, algorithm, privPEM, pubPEM string) error {
	_, err := k.pool.Exec(ctx, `INSERT INTO signing_keys (kid, algorithm, private_key, public_key) VALUES ($1, $2, $3, $4)`, kid, algorithm, privPEM, pubPEM)
	return err
}

func (k *PostgresKeyStore) AllPublicKeys(ctx context.Context) (map[string]string, error) {
	rows, err := k.pool.Query(ctx, `SELECT kid, public_key FROM signing_keys`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var kid, pub string
		if err := rows.Scan(&kid, &pub); err != nil {
			return nil, err
		}
		result[kid] = pub
	}
	return result, rows.Err()
}

// GenerateAndSaveKey is called once, at first-ever startup, when
// ActiveKey returns ErrNotFound — it creates the very first signing key.
func GenerateAndSaveKey(ctx context.Context, store domain.KeyStore, kid string) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	return store.Save(ctx, kid, "RS256", string(privPEM), string(pubPEM))
}
