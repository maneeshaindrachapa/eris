package domain

// Client is an application registered to use Eris. ClientType is what
// decides whether ClientSecretHash is meaningful at all: a public client
// (SPA, mobile app) has no secret — PKCE is its only proof of identity at
// the /token step, which is why AuthorizationCode carries a code challenge
// regardless of client type.
type Client struct {
	ID                string
	Name              string
	Type              ClientType
	ClientSecretHash  string // empty for public clients
	RedirectURIs      []string
	AllowedScopes     []string
	AllowedGrantTypes []GrantType
}

type ClientType string

const (
	ClientTypeConfidential ClientType = "confidential"
	ClientTypePublic       ClientType = "public"
)

type GrantType string

const (
	GrantTypeAuthorizationCode GrantType = "authorization_code"
	GrantTypeRefreshToken      GrantType = "refresh_token"
)

// HasRedirectURI does an EXACT match, deliberately — prefix or substring
// matching on redirect URIs is a classic open-redirect vulnerability in
// OAuth implementations.
func (c Client) HasRedirectURI(uri string) bool {
	for _, u := range c.RedirectURIs {
		if u == uri {
			return true
		}
	}
	return false
}

// SupportsScopes checks every requested scope is in the client's allow-list.
func (c Client) SupportsScopes(requested []string) bool {
	allowed := make(map[string]bool, len(c.AllowedScopes))
	for _, s := range c.AllowedScopes {
		allowed[s] = true
	}
	for _, s := range requested {
		if !allowed[s] {
			return false
		}
	}
	return true
}
