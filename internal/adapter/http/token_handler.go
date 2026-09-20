package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type TokenHandler struct {
	svc domain.TokenService
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope"`
}

func (h *TokenHandler) Token(c echo.Context) error {
	grantType := c.FormValue("grant_type")

	var result domain.TokenResult
	var err error

	switch grantType {
	case "authorization_code":
		result, err = h.svc.ExchangeAuthorizationCode(c.Request().Context(),
			c.FormValue("code"), c.FormValue("redirect_uri"), c.FormValue("code_verifier"),
			c.FormValue("client_id"), c.FormValue("client_secret"))
	case "refresh_token":
		result, err = h.svc.RefreshAccessToken(c.Request().Context(),
			c.FormValue("refresh_token"), c.FormValue("client_id"), c.FormValue("client_secret"))
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
	}

	if err != nil {
		return c.JSON(tokenErrorStatus(err), map[string]string{"error": tokenErrorCode(err), "error_description": err.Error()})
	}

	return c.JSON(http.StatusOK, tokenResponse{
		AccessToken: result.AccessToken, TokenType: "Bearer", ExpiresIn: result.ExpiresIn,
		RefreshToken: result.RefreshToken, IDToken: result.IDToken,
	})
}

func tokenErrorCode(err error) string {
	switch err.(type) {
	case domain.InvalidClientError:
		return "invalid_client"
	case domain.PKCEMismatchError, domain.InvalidGrantError:
		return "invalid_grant"
	default:
		return "server_error"
	}
}

func tokenErrorStatus(err error) int {
	switch err.(type) {
	case domain.InvalidClientError:
		return http.StatusUnauthorized
	case domain.PKCEMismatchError, domain.InvalidGrantError:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
