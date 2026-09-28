package application

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type tokenService struct {
	clients    domain.ClientRepository
	codes      domain.AuthCodeStore
	refresh    domain.RefreshTokenStore
	issuer     domain.TokenIssuer
	refreshTTL time.Duration
}

func NewTokenService(clients domain.ClientRepository, codes domain.AuthCodeStore, refresh domain.RefreshTokenStore, issuer domain.TokenIssuer) *tokenService {
	return &tokenService{clients: clients, codes: codes, refresh: refresh, issuer: issuer, refreshTTL: 30 * 24 * time.Hour}
}

func (s *tokenService) ExchangeAuthorizationCode(ctx context.Context, code, redirectURI, codeVerifier, clientID, clientSecret string) (domain.TokenResult, error) {
	client, err := s.authenticateClient(ctx, clientID, clientSecret)
	if err != nil {
		return domain.TokenResult{}, err
	}

	// Atomic consume — see the AuthCodeStore implementation notes. A reused
	// or expired code fails here with no further checks needed.
	ac, err := s.codes.Consume(ctx, code, time.Now())
	if err != nil {
		return domain.TokenResult{}, err
	}
	if ac.ClientID != client.ID {
		return domain.TokenResult{}, domain.ErrorInvalidGrant("code was not issued to this client")
	}
	// Re-checked here deliberately, even though /authorize already checked
	// it once — RFC 6749 requires it be validated again at redemption.
	if ac.RedirectURI != redirectURI {
		return domain.TokenResult{}, domain.ErrorInvalidGrant("redirect_uri does not match")
	}

	// The actual PKCE proof: recompute SHA256(code_verifier) and compare
	// against what was stored at /authorize time. Constant-time compare —
	// this is a security-sensitive equality check.
	challenge := computeS256Challenge(codeVerifier)
	if subtle.ConstantTimeCompare([]byte(challenge), []byte(ac.CodeChallenge)) != 1 {
		return domain.TokenResult{}, domain.ErrorPKCEMismatch("code_verifier does not match code_challenge")
	}

	return s.issueTokens(ctx, ac.UserID, client.ID, ac.Scopes)
}

func (s *tokenService) RefreshAccessToken(ctx context.Context, refreshTokenValue, clientID, clientSecret string) (domain.TokenResult, error) {
	client, err := s.authenticateClient(ctx, clientID, clientSecret)
	if err != nil {
		return domain.TokenResult{}, err
	}

	hash := hashToken(refreshTokenValue)
	rt, err := s.refresh.FindByTokenHash(ctx, hash)
	if err != nil {
		return domain.TokenResult{}, domain.ErrorInvalidGrant("unknown refresh token")
	}
	// A revoked token being presented again is exactly the theft signal
	// rotation is designed to catch — this is where you'd also revoke the
	// whole chain (ReplacedBy) in a hardened version.
	if rt.IsRevoked() || rt.ClientID != client.ID {
		return domain.TokenResult{}, domain.ErrorInvalidGrant("refresh token is no longer valid")
	}
	if time.Now().After(rt.ExpiresAt) {
		return domain.TokenResult{}, domain.ErrorInvalidGrant("refresh token expired")
	}

	if err := s.refresh.Revoke(ctx, rt.ID, time.Now()); err != nil { // rotation: old one dies here
		return domain.TokenResult{}, err
	}
	return s.issueTokens(ctx, rt.UserID, client.ID, rt.Scopes)
}

func (s *tokenService) issueTokens(ctx context.Context, userID, clientID string, scopes []string) (domain.TokenResult, error) {
	access, err := s.issuer.IssueAccessToken(ctx, userID, clientID, scopes)
	if err != nil {
		return domain.TokenResult{}, err
	}

	rawRefresh := randomOpaqueValue(32)
	newRT := domain.RefreshToken{ID: randomOpaqueValue(16), UserID: userID, ClientID: clientID, TokenHash: hashToken(rawRefresh), Scopes: scopes, ExpiresAt: time.Now().Add(s.refreshTTL)}
	if err := s.refresh.Save(ctx, newRT); err != nil {
		return domain.TokenResult{}, err
	}

	result := domain.TokenResult{AccessToken: access.Value, RefreshToken: rawRefresh, ExpiresIn: int(time.Until(access.ExpiresAt).Seconds()), Scopes: scopes}
	if containsScope(scopes, "openid") {
		idToken, err := s.issuer.IssueIDToken(ctx, userID, clientID, "")
		if err == nil {
			result.IDToken = idToken
		}
	}
	return result, nil
}

func (s *tokenService) authenticateClient(ctx context.Context, clientID, clientSecret string) (domain.Client, error) {
	client, err := s.clients.FindByID(ctx, clientID)
	if err != nil {
		return domain.Client{}, domain.ErrorInvalidClient("unknown client")
	}
	if client.Type == domain.ClientTypeConfidential {
		if bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(clientSecret)) != nil {
			return domain.Client{}, domain.ErrorInvalidClient("invalid client secret")
		}
	}
	// Public clients present no secret at all — PKCE (checked by the caller
	// for the code exchange path) is their proof instead.
	return client, nil
}

func computeS256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func containsScope(scopes []string, target string) bool {
	for _, s := range scopes {
		if s == target {
			return true
		}
	}
	return false
}
