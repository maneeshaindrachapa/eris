package memory

import (
	"context"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type ClientRepository struct {
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
	c, ok := r.byID[id]
	if !ok {
		return domain.Client{}, domain.ErrorNotFound("Client by id %s not found", id)
	}
	return c, nil
}
