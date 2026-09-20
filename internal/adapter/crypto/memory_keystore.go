package crypto

import (
	"context"
	"sync"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type memoryKey struct {
	privateKeyPEM string
	publicKeyPEM  string
}

type MemoryKeyStore struct {
	mu        sync.RWMutex
	activeKID string
	keys      map[string]memoryKey
}

func NewMemoryKeyStore() *MemoryKeyStore {
	return &MemoryKeyStore{keys: make(map[string]memoryKey)}
}

func (k *MemoryKeyStore) ActiveKey(ctx context.Context) (string, string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.activeKID == "" {
		return "", "", domain.NotFoundError{Message: "no active signing key"}
	}
	key := k.keys[k.activeKID]
	return k.activeKID, key.privateKeyPEM, nil
}

func (k *MemoryKeyStore) Save(ctx context.Context, kid, algorithm, privateKeyPEM, publicKeyPEM string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.keys[kid] = memoryKey{privateKeyPEM: privateKeyPEM, publicKeyPEM: publicKeyPEM}
	k.activeKID = kid
	return nil
}

func (k *MemoryKeyStore) AllPublicKeys(ctx context.Context) (map[string]string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	keys := make(map[string]string, len(k.keys))
	for kid, key := range k.keys {
		keys[kid] = key.publicKeyPEM
	}
	return keys, nil
}
