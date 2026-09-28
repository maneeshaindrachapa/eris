package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type ClientRepository struct {
	mu   sync.RWMutex
	byID map[string]domain.Client
}

func NewClientRepository() *ClientRepository {
	return &ClientRepository{
		byID: map[string]domain.Client{
			"spa-client": {
				ID:                "spa-client",
				Name:              "Demo SPA",
				Type:              domain.ClientTypePublic,
				RedirectURIs:      []string{"http://localhost:3000/callback"},
				AllowedScopes:     []string{"openid", "profile", "email"},
				AllowedGrantTypes: []domain.GrantType{domain.GrantTypeAuthorizationCode, domain.GrantTypeRefreshToken},
			},
		},
	}
}

func (r *ClientRepository) FindByID(ctx context.Context, id string) (domain.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byID[id]
	if !ok {
		return domain.Client{}, domain.ErrorNotFound("Client by id %s not found", id)
	}
	return c, nil
}

func (r *ClientRepository) List(ctx context.Context) ([]domain.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	clients := make([]domain.Client, 0, len(r.byID))
	for _, client := range r.byID {
		clients = append(clients, client)
	}
	sort.Slice(clients, func(i, j int) bool { return clients[i].Name < clients[j].Name })
	return clients, nil
}

func (r *ClientRepository) Create(ctx context.Context, client domain.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byID[client.ID]; exists {
		return domain.ErrorInvalidGrant("client ID already exists")
	}
	r.byID[client.ID] = client
	return nil
}

func (r *ClientRepository) Update(ctx context.Context, client domain.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byID[client.ID]; !exists {
		return domain.ErrorNotFound("client not found")
	}
	r.byID[client.ID] = client
	return nil
}
