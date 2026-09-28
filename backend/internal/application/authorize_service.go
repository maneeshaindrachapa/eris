package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type authorizeService struct {
	clients  domain.ClientRepository
	users    domain.UserRepository
	codes    domain.AuthCodeStore
	sessions domain.SessionStore
	consents domain.ConsentStore
	codeTTL  time.Duration
}

func NewAuthorizeService(
	clients domain.ClientRepository,
	users domain.UserRepository,
	codes domain.AuthCodeStore,
	sessions domain.SessionStore,
	consents domain.ConsentStore,
) *authorizeService {
	return &authorizeService{clients: clients, users: users, codes: codes, sessions: sessions, consents: consents, codeTTL: 90 * time.Second}
}

// Authorize runs the checklist from Step 4, in order. Anything wrong in
// steps 1-2 is NOT safe to redirect anywhere (we don't have a trusted
// redirect_uri yet) — the caller (HTTP handler) must show these as direct
// errors, never as a redirect.
func (s *authorizeService) Authorize(ctx context.Context, req domain.AuthorizeRequest) (domain.AuthorizeResult, error) {
	client, err := s.clients.FindByID(ctx, req.ClientID)
	if err != nil {
		log.Printf("DEBUG: client lookup failed for client_id=%q: %v", req.ClientID, err)
		return domain.AuthorizeResult{}, domain.InvalidClientError{Message: "unknown client_id"}
	}
	log.Printf("DEBUG: found client %+v", client)

	if !client.HasRedirectURI(req.RedirectURI) {
		log.Printf("DEBUG: redirect_uri mismatch — got %q, registered: %v", req.RedirectURI, client.RedirectURIs)
		return domain.AuthorizeResult{}, domain.InvalidClientError{Message: "redirect_uri not registered for this client"}
	}

	// From here on, errors CAN be sent back via redirect with an error code,
	// per RFC 6749 §4.1.2.1 — the handler is responsible for that framing,
	// this layer just returns domain errors either way.
	if !client.SupportsScopes(req.Scopes) {
		return domain.AuthorizeResult{}, domain.ErrorInvalidScope("scope not allowed for this client")
	}
	if req.CodeChallenge == "" || req.CodeChallengeMethod != "S256" {
		return domain.AuthorizeResult{}, domain.ErrorInvalidGrant("PKCE (S256) is required")
	}

	// No session cookie at all -> show login.
	if req.SessionID == "" {
		return domain.AuthorizeResult{NeedsLogin: true}, nil
	}
	sess, err := s.sessions.Find(ctx, req.SessionID)
	if err != nil || sess.IsExpired(time.Now()) {
		return domain.AuthorizeResult{NeedsLogin: true}, nil
	}

	// Have a session — check whether this client+scope combo was already
	// approved, so returning users don't get re-prompted every time.
	consent, err := s.consents.Find(ctx, sess.UserID, req.ClientID)
	if err != nil || !containsAll(consent.Scopes, req.Scopes) {
		return domain.AuthorizeResult{NeedsConsent: true}, nil
	}

	code, err := s.issueCode(ctx, sess.UserID, req)
	if err != nil {
		return domain.AuthorizeResult{}, err
	}
	return domain.AuthorizeResult{Code: code}, nil
}

// Login verifies credentials and creates a new Eris browser session. This
// is deliberately a separate method from Authorize — the login FORM is an
// HTTP/rendering concern, but verifying the password and minting a session
// is business logic that belongs here, not in the handler.
func (s *authorizeService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrorUnauthorized("invalid email or password")
	}
	if user.Status != domain.UserStatusActive {
		return "", domain.ErrorUnauthorized("account is locked")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", domain.ErrorUnauthorized("invalid email or password")
	}

	sessionID := randomOpaqueValue(32)
	sess := domain.Session{ID: sessionID, UserID: user.ID, ExpiresAt: time.Now().Add(24 * time.Hour)}
	if err := s.sessions.Save(ctx, sess); err != nil {
		return "", err
	}
	return sessionID, nil
}

// ApproveConsent records that the session's user approved this client for
// these scopes, then issues the authorization code immediately — the
// handler doesn't need to call Authorize a third time.
func (s *authorizeService) ApproveConsent(ctx context.Context, req domain.AuthorizeRequest, approvedScopes []string) (string, error) {
	sess, err := s.sessions.Find(ctx, req.SessionID)
	if err != nil {
		return "", domain.ErrorUnauthorized("session expired")
	}
	if err := s.consents.Save(ctx, domain.Consent{UserID: sess.UserID, ClientID: req.ClientID, Scopes: approvedScopes, GrantedAt: time.Now()}); err != nil {
		return "", err
	}
	return s.issueCode(ctx, sess.UserID, req)
}

func (s *authorizeService) issueCode(ctx context.Context, userID string, req domain.AuthorizeRequest) (string, error) {
	code := randomOpaqueValue(32)
	ac := domain.AuthorizationCode{
		Code: code, UserID: userID, ClientID: req.ClientID, RedirectURI: req.RedirectURI,
		Scopes: req.Scopes, CodeChallenge: req.CodeChallenge, CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt: time.Now().Add(s.codeTTL),
	}
	if err := s.codes.Save(ctx, ac); err != nil {
		return "", err
	}
	return code, nil
}

func randomOpaqueValue(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func containsAll(have, want []string) bool {
	set := make(map[string]bool, len(have))
	for _, s := range have {
		set[s] = true
	}
	for _, s := range want {
		if !set[s] {
			return false
		}
	}
	return true
}
