package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type adminService struct{ clients domain.ClientRepository }

func NewAdminService(clients domain.ClientRepository) *adminService {
	return &adminService{clients: clients}
}

func (s *adminService) ListClients(ctx context.Context) ([]domain.Client, error) {
	return s.clients.List(ctx)
}

func (s *adminService) CreateClient(ctx context.Context, input domain.AdminClientInput) (domain.AdminClientResult, error) {
	if err := validateClientInput(input); err != nil {
		return domain.AdminClientResult{}, err
	}
	client := domain.Client{
		ID: "client-" + randomHex(12), Name: strings.TrimSpace(input.Name), Type: input.Type,
		RedirectURIs: input.RedirectURIs, AllowedScopes: input.AllowedScopes, AllowedGrantTypes: input.AllowedGrantTypes,
	}
	secret := ""
	if input.Type == domain.ClientTypeConfidential {
		secret = randomSecret(32)
		hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
		if err != nil {
			return domain.AdminClientResult{}, fmt.Errorf("hash client secret: %w", err)
		}
		client.ClientSecretHash = string(hash)
	}
	if err := s.clients.Create(ctx, client); err != nil {
		return domain.AdminClientResult{}, err
	}
	return domain.AdminClientResult{Client: client, ClientSecret: secret}, nil
}

func (s *adminService) UpdateClient(ctx context.Context, id string, input domain.AdminClientInput) (domain.Client, error) {
	if err := validateClientInput(input); err != nil {
		return domain.Client{}, err
	}
	client, err := s.clients.FindByID(ctx, id)
	if err != nil {
		return domain.Client{}, err
	}
	client.Name = strings.TrimSpace(input.Name)
	client.RedirectURIs = input.RedirectURIs
	client.AllowedScopes = input.AllowedScopes
	client.AllowedGrantTypes = input.AllowedGrantTypes
	if err := s.clients.Update(ctx, client); err != nil {
		return domain.Client{}, err
	}
	return client, nil
}

func validateClientInput(input domain.AdminClientInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return domain.ErrorInvalidGrant("client name is required")
	}
	if input.Type != domain.ClientTypePublic && input.Type != domain.ClientTypeConfidential {
		return domain.ErrorInvalidGrant("client type must be public or confidential")
	}
	if len(input.RedirectURIs) == 0 {
		return domain.ErrorInvalidGrant("at least one redirect URI is required")
	}
	for _, raw := range input.RedirectURIs {
		u, err := url.ParseRequestURI(raw)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return domain.ErrorInvalidGrant("invalid redirect URI %q", raw)
		}
	}
	if len(input.AllowedScopes) == 0 {
		return domain.ErrorInvalidGrant("at least one scope is required")
	}
	if len(input.AllowedGrantTypes) == 0 {
		return domain.ErrorInvalidGrant("at least one grant type is required")
	}
	return nil
}

func randomHex(size int) string {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value)
}
func randomSecret(size int) string {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(value)
}
