package memory

import (
	"context"
	"sync"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type ConsentStore struct {
	mu       sync.RWMutex
	consents map[string]domain.Consent // keyed by userID+"|"+clientID
}

func NewConsentStore() *ConsentStore {
	return &ConsentStore{consents: make(map[string]domain.Consent)}
}

func key(userID, clientID string) string { return userID + "|" + clientID }

func (s *ConsentStore) Find(ctx context.Context, userID, clientID string) (domain.Consent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.consents[key(userID, clientID)]
	if !ok {
		return domain.Consent{}, domain.NotFoundError{Message: "no consent found"}
	}
	return c, nil
}

func (s *ConsentStore) Save(ctx context.Context, c domain.Consent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consents[key(c.UserID, c.ClientID)] = c
	return nil
}
