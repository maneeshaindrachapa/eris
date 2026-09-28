package memory

import (
	"context"
	"sync"
	"time"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type RefreshTokenStore struct {
	mu     sync.Mutex
	tokens map[string]domain.RefreshToken
}

func NewRefreshTokenStore() *RefreshTokenStore {
	return &RefreshTokenStore{tokens: make(map[string]domain.RefreshToken)}
}

func (s *RefreshTokenStore) Save(ctx context.Context, rt domain.RefreshToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[rt.TokenHash] = rt
	return nil
}

func (s *RefreshTokenStore) FindByTokenHash(ctx context.Context, hash string) (domain.RefreshToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rt, ok := s.tokens[hash]
	if !ok {
		return domain.RefreshToken{}, domain.NotFoundError{Message: "refresh token not found"}
	}
	return rt, nil
}

func (s *RefreshTokenStore) Revoke(ctx context.Context, id string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for hash, rt := range s.tokens {
		if rt.ID == id {
			rt.RevokedAt = &now
			s.tokens[hash] = rt
			return nil
		}
	}
	return domain.NotFoundError{Message: "refresh token not found"}
}
