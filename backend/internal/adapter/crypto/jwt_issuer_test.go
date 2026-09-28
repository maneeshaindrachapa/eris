package crypto

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

func TestNewJWTIssuerGeneratesAndSavesMissingKey(t *testing.T) {
	store := &stubKeyStore{activeErr: domain.NotFoundError{Message: "missing"}}

	issuer, err := NewJWTIssuer(context.Background(), store, "https://issuer.example.com")
	if err != nil {
		t.Fatalf("NewJWTIssuer() error = %v", err)
	}
	if issuer == nil {
		t.Fatal("NewJWTIssuer() returned nil issuer")
	}
	if store.saveCalls != 1 {
		t.Fatalf("Save() calls = %d, want 1", store.saveCalls)
	}
	if store.algorithm != "RS256" {
		t.Fatalf("saved algorithm = %q, want RS256", store.algorithm)
	}
	if store.kid == "" || store.privateKeyPEM == "" || store.publicKeyPEM == "" {
		t.Fatal("generated key data was not saved")
	}
	assertPrivateKeyPEM(t, store.privateKeyPEM)
	assertPublicKeyPEM(t, store.publicKeyPEM)
}

func TestNewJWTIssuerPropagatesKeyStoreErrors(t *testing.T) {
	t.Run("active key", func(t *testing.T) {
		wantErr := errors.New("read key")
		_, err := NewJWTIssuer(context.Background(), &stubKeyStore{activeErr: wantErr}, "issuer")
		if !errors.Is(err, wantErr) {
			t.Fatalf("NewJWTIssuer() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("save generated key", func(t *testing.T) {
		wantErr := errors.New("save key")
		store := &stubKeyStore{
			activeErr: domain.NotFoundError{Message: "missing"},
			saveErr:   wantErr,
		}
		_, err := NewJWTIssuer(context.Background(), store, "issuer")
		if !errors.Is(err, wantErr) {
			t.Fatalf("NewJWTIssuer() error = %v, want %v", err, wantErr)
		}
	})
}

func TestNewJWTIssuerRejectsInvalidPrivateKey(t *testing.T) {
	tests := []struct {
		name string
		pem  string
	}{
		{name: "not PEM", pem: "not-a-private-key"},
		{name: "not PKCS1", pem: string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: []byte("invalid")}))},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewJWTIssuer(context.Background(), &stubKeyStore{kid: "key-1", privateKeyPEM: test.pem}, "issuer")
			if err == nil {
				t.Fatal("NewJWTIssuer() error = nil, want key parsing error")
			}
		})
	}
}

func TestIssueAccessTokenProducesValidClaims(t *testing.T) {
	privateKey, privateKeyPEM, publicKeyPEM := makeRSAKey(t)
	store := &stubKeyStore{kid: "key-123", privateKeyPEM: privateKeyPEM, publicKeyPEM: publicKeyPEM}
	issuer, err := NewJWTIssuer(context.Background(), store, "https://issuer.example.com")
	if err != nil {
		t.Fatalf("NewJWTIssuer() error = %v", err)
	}

	before := time.Now()
	accessToken, err := issuer.IssueAccessToken(context.Background(), "user-1", "spa-client", []string{"openid", "profile"})
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}
	after := time.Now()

	claims := parseAndVerifyToken(t, accessToken.Value, &privateKey.PublicKey, "https://issuer.example.com", "spa-client")
	if claims["sub"] != "user-1" {
		t.Fatalf("sub claim = %#v", claims["sub"])
	}
	if got := stringSliceClaim(t, claims, "scope"); !reflect.DeepEqual(got, []string{"openid", "profile"}) {
		t.Fatalf("scope claim = %#v", got)
	}
	if accessToken.ExpiresAt.Before(before.Add(15*time.Minute)) || accessToken.ExpiresAt.After(after.Add(15*time.Minute)) {
		t.Fatalf("ExpiresAt = %s, want approximately 15 minutes", accessToken.ExpiresAt)
	}
}

func TestIssueIDTokenHandlesNonce(t *testing.T) {
	privateKey, privateKeyPEM, publicKeyPEM := makeRSAKey(t)
	issuer, err := NewJWTIssuer(context.Background(), &stubKeyStore{
		kid: "key-123", privateKeyPEM: privateKeyPEM, publicKeyPEM: publicKeyPEM,
	}, "https://issuer.example.com")
	if err != nil {
		t.Fatalf("NewJWTIssuer() error = %v", err)
	}

	t.Run("included", func(t *testing.T) {
		raw, err := issuer.IssueIDToken(context.Background(), "user-1", "spa-client", "nonce-123")
		if err != nil {
			t.Fatalf("IssueIDToken() error = %v", err)
		}
		claims := parseAndVerifyToken(t, raw, &privateKey.PublicKey, "https://issuer.example.com", "spa-client")
		if claims["nonce"] != "nonce-123" {
			t.Fatalf("nonce claim = %#v", claims["nonce"])
		}
	})

	t.Run("omitted when empty", func(t *testing.T) {
		raw, err := issuer.IssueIDToken(context.Background(), "user-1", "spa-client", "")
		if err != nil {
			t.Fatalf("IssueIDToken() error = %v", err)
		}
		claims := parseAndVerifyToken(t, raw, &privateKey.PublicKey, "https://issuer.example.com", "spa-client")
		if _, exists := claims["nonce"]; exists {
			t.Fatalf("nonce claim unexpectedly present: %#v", claims["nonce"])
		}
	})
}

type stubKeyStore struct {
	kid           string
	algorithm     string
	privateKeyPEM string
	publicKeyPEM  string
	activeErr     error
	saveErr       error
	saveCalls     int
}

func (s *stubKeyStore) ActiveKey(ctx context.Context) (string, string, error) {
	if s.activeErr != nil {
		return "", "", s.activeErr
	}
	return s.kid, s.privateKeyPEM, nil
}

func (s *stubKeyStore) Save(ctx context.Context, kid, algorithm, privateKeyPEM, publicKeyPEM string) error {
	s.saveCalls++
	if s.saveErr != nil {
		return s.saveErr
	}
	s.kid = kid
	s.algorithm = algorithm
	s.privateKeyPEM = privateKeyPEM
	s.publicKeyPEM = publicKeyPEM
	s.activeErr = nil
	return nil
}

func (s *stubKeyStore) AllPublicKeys(ctx context.Context) (map[string]string, error) {
	return map[string]string{s.kid: s.publicKeyPEM}, nil
}

func makeRSAKey(t *testing.T) (*rsa.PrivateKey, string, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("x509.MarshalPKIXPublicKey() error = %v", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER})
	return privateKey, string(privateKeyPEM), string(publicKeyPEM)
}

func parseAndVerifyToken(t *testing.T, raw string, publicKey *rsa.PublicKey, issuer, audience string) jwt.MapClaims {
	t.Helper()
	token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodRS256 {
			t.Fatalf("signing method = %s, want RS256", token.Method.Alg())
		}
		if token.Header["kid"] != "key-123" {
			t.Fatalf("kid header = %#v", token.Header["kid"])
		}
		return publicKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer(issuer), jwt.WithAudience(audience))
	if err != nil {
		t.Fatalf("jwt.Parse() error = %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("claims type = %T", token.Claims)
	}
	return claims
}

func stringSliceClaim(t *testing.T, claims jwt.MapClaims, name string) []string {
	t.Helper()
	values, ok := claims[name].([]any)
	if !ok {
		t.Fatalf("%s claim type = %T", name, claims[name])
	}
	result := make([]string, len(values))
	for i, value := range values {
		result[i], ok = value.(string)
		if !ok {
			t.Fatalf("%s[%d] type = %T", name, i, value)
		}
	}
	return result
}

func assertPrivateKeyPEM(t *testing.T, value string) {
	t.Helper()
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		t.Fatal("private key is not PEM")
	}
	if _, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		t.Fatalf("x509.ParsePKCS1PrivateKey() error = %v", err)
	}
}

func assertPublicKeyPEM(t *testing.T, value string) {
	t.Helper()
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		t.Fatal("public key is not PEM")
	}
	if _, err := x509.ParsePKIXPublicKey(block.Bytes); err != nil {
		t.Fatalf("x509.ParsePKIXPublicKey() error = %v", err)
	}
}
