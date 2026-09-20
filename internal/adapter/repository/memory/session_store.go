package memory

import (
	"context"
	"sync"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]domain.Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]domain.Session)}
}

func (s *SessionStore) Save(ctx context.Context, sess domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.ID] = sess
	return nil
}

func (s *SessionStore) Find(ctx context.Context, id string) (domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return domain.Session{}, domain.NotFoundError{Message: "session not found"}
	}
	return sess, nil
}
