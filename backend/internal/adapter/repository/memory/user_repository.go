package memory

import (
	"context"
	"sync"

	"github.com/maneeshaindrachapa/eris/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	mu    sync.RWMutex
	byID  map[string]domain.User
	email map[string]string
}

func NewUserRepository() *UserRepository {
	repo := &UserRepository{
		byID:  make(map[string]domain.User),
		email: make(map[string]string),
	}
	repo.seed()
	return repo
}

func (r *UserRepository) seed() {
	hash := func(pw string) string {
		h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		return string(h)
	}

	users := []domain.User{
		{ID: "user-1", Email: "alice@example.com", PasswordHash: hash("correct-horse"), Status: domain.UserStatusActive, EmailVerified: true},
		{ID: "user-2", Email: "bob@example.com", PasswordHash: hash("battery-staple"), Status: domain.UserStatusActive, EmailVerified: true},
	}

	for _, u := range users {
		r.byID[u.ID] = u
		r.email[u.Email] = u.ID
	}
}

func (r *UserRepository) FindByID(c context.Context, id string) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.byID[id]
	if !ok {
		return domain.User{}, domain.ErrorNotFound("User by id %s not found", id)
	}
	return u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.email[email]
	if !ok {
		return domain.User{}, domain.ErrorNotFound("User by E-Mail %s not found", email)
	}
	return r.byID[id], nil
}
