package crypto

import (
	"context"
	"errors"
	"testing"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

func TestMemoryKeyStoreStartsEmpty(t *testing.T) {
	store := NewMemoryKeyStore()

	_, _, err := store.ActiveKey(context.Background())
	var notFound domain.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("ActiveKey() error = %v, want NotFoundError", err)
	}
}

func TestJWTIssuerGeneratesAndReusesMemoryKey(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryKeyStore()

	first, err := NewJWTIssuer(ctx, store, "http://localhost:8080")
	if err != nil {
		t.Fatalf("first NewJWTIssuer() error = %v", err)
	}
	second, err := NewJWTIssuer(ctx, store, "http://localhost:8080")
	if err != nil {
		t.Fatalf("second NewJWTIssuer() error = %v", err)
	}

	if first.kid != second.kid {
		t.Fatalf("issuer key IDs differ: %q != %q", first.kid, second.kid)
	}
	if first.privateKey.D.Cmp(second.privateKey.D) != 0 {
		t.Fatal("issuer did not reuse the stored private key")
	}

	publicKeys, err := store.AllPublicKeys(ctx)
	if err != nil {
		t.Fatalf("AllPublicKeys() error = %v", err)
	}
	if len(publicKeys) != 1 || publicKeys[first.kid] == "" {
		t.Fatalf("AllPublicKeys() = %#v", publicKeys)
	}
}

func TestMemoryKeyStoreImplementsDomainPort(t *testing.T) {
	var _ domain.KeyStore = NewMemoryKeyStore()
}

func TestMemoryKeyStoreReturnsPublicKeySnapshot(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryKeyStore()
	if err := store.Save(ctx, "key-1", "RS256", "private-1", "public-1"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := store.Save(ctx, "key-2", "RS256", "private-2", "public-2"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	kid, privateKey, err := store.ActiveKey(ctx)
	if err != nil {
		t.Fatalf("ActiveKey() error = %v", err)
	}
	if kid != "key-2" || privateKey != "private-2" {
		t.Fatalf("ActiveKey() = (%q, %q), want key-2/private-2", kid, privateKey)
	}

	publicKeys, err := store.AllPublicKeys(ctx)
	if err != nil {
		t.Fatalf("AllPublicKeys() error = %v", err)
	}
	publicKeys["key-1"] = "changed"
	publicKeysAgain, err := store.AllPublicKeys(ctx)
	if err != nil {
		t.Fatalf("second AllPublicKeys() error = %v", err)
	}
	if publicKeysAgain["key-1"] != "public-1" {
		t.Fatalf("stored public key changed through returned map: %q", publicKeysAgain["key-1"])
	}
}
