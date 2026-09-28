package memory

import (
	"context"
	"sync"
	"time"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type AuthCodeStore struct {
	mu    sync.Mutex
	codes map[string]domain.AuthorizationCode
}

func NewAuthCodeStore() *AuthCodeStore {
	return &AuthCodeStore{codes: make(map[string]domain.AuthorizationCode)}
}

func (s *AuthCodeStore) Save(ctx context.Context, code domain.AuthorizationCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[code.Code] = code
	return nil
}

func (s *AuthCodeStore) Consume(ctx context.Context, code string, now time.Time) (domain.AuthorizationCode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ac, ok := s.codes[code]
	if !ok {
		return domain.AuthorizationCode{}, domain.ErrorInvalidGrant("unknown authorization code")
	}
	if ac.IsConsumed() {
		return domain.AuthorizationCode{}, domain.ErrorInvalidGrant("authorization code already used")
	}
	if ac.IsExpired(now) {
		return domain.AuthorizationCode{}, domain.ErrorInvalidGrant("authorization code expired")
	}

	ac.ConsumedAt = &now
	s.codes[code] = ac
	return ac, nil
}
