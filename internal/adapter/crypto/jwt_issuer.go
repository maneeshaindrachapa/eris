package crypto

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type JWTIssuer struct {
	kid        string
	privateKey *rsa.PrivateKey
	issuer     string
	accessTTL  time.Duration
}

// NewJWTIssuer now takes a KeyStore instead of generating a key itself.
// If no key exists yet (first-ever boot), it generates and persists one —
// after that, every restart loads the SAME key, so tokens signed before a
// restart stay verifiable.
func NewJWTIssuer(ctx context.Context, store domain.KeyStore, issuerURL string) (*JWTIssuer, error) {
	kid, privPEM, err := store.ActiveKey(ctx)
	if err != nil {
		var notFound domain.NotFoundError
		if !errors.As(err, &notFound) {
			return nil, err
		}
		kid = "key-" + time.Now().Format("2006-01-02")
		if err := GenerateAndSaveKey(ctx, store, kid); err != nil {
			return nil, err
		}
		_, privPEM, err = store.ActiveKey(ctx)
		if err != nil {
			return nil, err
		}
	}

	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("decode signing key PEM")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &JWTIssuer{kid: kid, privateKey: key, issuer: issuerURL, accessTTL: 15 * time.Minute}, nil
}

func (j *JWTIssuer) IssueAccessToken(ctx context.Context, userID, clientID string, scopes []string) (domain.AccessToken, error) {
	now := time.Now()
	exp := now.Add(j.accessTTL)
	claims := jwt.MapClaims{"iss": j.issuer, "sub": userID, "aud": clientID, "scope": scopes, "iat": now.Unix(), "exp": exp.Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = j.kid
	signed, err := token.SignedString(j.privateKey)
	if err != nil {
		return domain.AccessToken{}, err
	}
	return domain.AccessToken{Value: signed, ExpiresAt: exp}, nil
}

func (j *JWTIssuer) IssueIDToken(ctx context.Context, userID, clientID, nonce string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{"iss": j.issuer, "sub": userID, "aud": clientID, "iat": now.Unix(), "exp": now.Add(j.accessTTL).Unix()}
	if nonce != "" {
		claims["nonce"] = nonce
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = j.kid
	return token.SignedString(j.privateKey)
}
