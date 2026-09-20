package http

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

const sessionCookieName = "eris_session"

type AuthorizeHandler struct {
	svc domain.AuthorizeService
}

// buildRequest reads the /authorize query params (or, for the follow-up
// POSTs, the same values re-submitted as hidden form fields) into the
// domain-level AuthorizeRequest. This is the ONLY place HTTP param names
// get translated into the shape the application layer expects.
func buildRequest(c echo.Context) domain.AuthorizeRequest {
	sessionID := ""
	if cookie, err := c.Cookie(sessionCookieName); err == nil {
		sessionID = cookie.Value
	}
	return domain.AuthorizeRequest{
		ClientID:            c.FormValue("client_id"),
		RedirectURI:         c.FormValue("redirect_uri"),
		Scopes:              strings.Fields(c.FormValue("scope")),
		State:               c.FormValue("state"),
		CodeChallenge:       c.FormValue("code_challenge"),
		CodeChallengeMethod: c.FormValue("code_challenge_method"),
		SessionID:           sessionID,
	}
}

// Start handles GET /authorize — the entry point of the flow.
func (h *AuthorizeHandler) Start(c echo.Context) error {
	req := buildRequest(c)
	result, err := h.svc.Authorize(c.Request().Context(), req)
	if err != nil {
		return renderAuthorizeError(c, req, err)
	}

	switch {
	case result.NeedsLogin:
		return c.HTML(http.StatusOK, loginFormHTML(req))
	case result.NeedsConsent:
		return c.HTML(http.StatusOK, consentFormHTML(req))
	default:
		return redirectWithCode(c, req.RedirectURI, result.Code, req.State)
	}
}

// Login handles POST /authorize/login — verifies credentials, sets the
// session cookie, then re-evaluates Authorize now that a session exists
// (which will very likely land on NeedsConsent next, unless already granted).
func (h *AuthorizeHandler) Login(c echo.Context) error {
	req := buildRequest(c)
	sessionID, err := h.svc.Login(c.Request().Context(), c.FormValue("email"), c.FormValue("password"))
	if err != nil {
		return c.HTML(http.StatusUnauthorized, loginFormHTML(req)+`<p class="error">Invalid email or password.</p>`)
	}

	c.SetCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, Expires: time.Now().Add(24 * time.Hour)})
	req.SessionID = sessionID

	result, err := h.svc.Authorize(c.Request().Context(), req)
	if err != nil {
		return renderAuthorizeError(c, req, err)
	}
	if result.NeedsConsent {
		return c.HTML(http.StatusOK, consentFormHTML(req))
	}
	return redirectWithCode(c, req.RedirectURI, result.Code, req.State)
}

// Consent handles POST /authorize/consent — the final step: records
// approval and redirects back to the client with the code.
func (h *AuthorizeHandler) Consent(c echo.Context) error {
	req := buildRequest(c)
	approved := c.Request().Form["approved_scope"] // checkboxes on the consent form

	code, err := h.svc.ApproveConsent(c.Request().Context(), req, approved)
	if err != nil {
		return renderAuthorizeError(c, req, err)
	}
	return redirectWithCode(c, req.RedirectURI, code, req.State)
}

func redirectWithCode(c echo.Context, redirectURI, code, state string) error {
	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("code", code)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return c.Redirect(http.StatusFound, u.String())
}

// renderAuthorizeError decides whether an error is safe to redirect back to
// the client (per RFC 6749 §4.1.2.1, only once client_id + redirect_uri are
// already validated) or must be shown directly, because we don't yet have
// a trusted place to send the browser.
func renderAuthorizeError(c echo.Context, req domain.AuthorizeRequest, err error) error {
	switch err.(type) {
	case domain.InvalidClientError:
		return c.String(http.StatusBadRequest, "invalid client_id or redirect_uri — request rejected") // NOT safe to redirect
	default:
		u, parseErr := url.Parse(req.RedirectURI)
		if parseErr != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		q := u.Query()
		q.Set("error", oauthErrorCode(err))
		q.Set("error_description", err.Error())
		q.Set("state", req.State)
		u.RawQuery = q.Encode()
		return c.Redirect(http.StatusFound, u.String())
	}
}

func oauthErrorCode(err error) string {
	switch err.(type) {
	case domain.InvalidScopeError:
		return "invalid_scope"
	case domain.InvalidGrantError:
		return "invalid_request"
	default:
		return "server_error"
	}
}

// Minimal inline forms — a real deployment renders proper templates, but
// the point here is the hidden-field round-trip, not styling.
func loginFormHTML(req domain.AuthorizeRequest) string {
	return `<form method="POST" action="/authorize/login">` + hiddenFields(req) + `
		<input name="email" type="email" placeholder="email">
		<input name="password" type="password" placeholder="password">
		<button type="submit">Log in</button></form>`
}

func consentFormHTML(req domain.AuthorizeRequest) string {
	var scopes strings.Builder
	for _, s := range req.Scopes {
		scopes.WriteString(`<label><input type="checkbox" name="approved_scope" value="` + s + `" checked> ` + s + `</label><br>`)
	}
	return `<form method="POST" action="/authorize/consent">` + hiddenFields(req) + scopes.String() + `
		<button type="submit">Approve</button></form>`
}

func hiddenFields(req domain.AuthorizeRequest) string {
	return `<input type="hidden" name="client_id" value="` + req.ClientID + `">` +
		`<input type="hidden" name="redirect_uri" value="` + req.RedirectURI + `">` +
		`<input type="hidden" name="scope" value="` + strings.Join(req.Scopes, " ") + `">` +
		`<input type="hidden" name="state" value="` + req.State + `">` +
		`<input type="hidden" name="code_challenge" value="` + req.CodeChallenge + `">` +
		`<input type="hidden" name="code_challenge_method" value="` + req.CodeChallengeMethod + `">`
}
