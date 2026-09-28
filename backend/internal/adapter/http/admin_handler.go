package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

type AdminHandler struct{ svc domain.AdminService }

type adminClientRequest struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	RedirectURIs      []string `json:"redirect_uris"`
	AllowedScopes     []string `json:"allowed_scopes"`
	AllowedGrantTypes []string `json:"allowed_grant_types"`
}

type adminClientResponse struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	RedirectURIs      []string `json:"redirect_uris"`
	AllowedScopes     []string `json:"allowed_scopes"`
	AllowedGrantTypes []string `json:"allowed_grant_types"`
	ClientSecret      string   `json:"client_secret,omitempty"`
}

func (h *AdminHandler) ListClients(c echo.Context) error {
	clients, err := h.svc.ListClients(c.Request().Context())
	if err != nil {
		return adminError(c, err)
	}
	response := make([]adminClientResponse, 0, len(clients))
	for _, client := range clients {
		response = append(response, clientResponse(client, ""))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *AdminHandler) CreateClient(c echo.Context) error {
	var request adminClientRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
	}
	result, err := h.svc.CreateClient(c.Request().Context(), request.input())
	if err != nil {
		return adminError(c, err)
	}
	return c.JSON(http.StatusCreated, clientResponse(result.Client, result.ClientSecret))
}

func (h *AdminHandler) UpdateClient(c echo.Context) error {
	var request adminClientRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
	}
	client, err := h.svc.UpdateClient(c.Request().Context(), c.Param("id"), request.input())
	if err != nil {
		return adminError(c, err)
	}
	return c.JSON(http.StatusOK, clientResponse(client, ""))
}

func (r adminClientRequest) input() domain.AdminClientInput {
	grants := make([]domain.GrantType, len(r.AllowedGrantTypes))
	for i, grant := range r.AllowedGrantTypes {
		grants[i] = domain.GrantType(grant)
	}
	return domain.AdminClientInput{Name: r.Name, Type: domain.ClientType(r.Type), RedirectURIs: r.RedirectURIs, AllowedScopes: r.AllowedScopes, AllowedGrantTypes: grants}
}

func clientResponse(client domain.Client, secret string) adminClientResponse {
	grants := make([]string, len(client.AllowedGrantTypes))
	for i, grant := range client.AllowedGrantTypes {
		grants[i] = string(grant)
	}
	return adminClientResponse{ID: client.ID, Name: client.Name, Type: string(client.Type), RedirectURIs: client.RedirectURIs, AllowedScopes: client.AllowedScopes, AllowedGrantTypes: grants, ClientSecret: secret}
}

func adminError(c echo.Context, err error) error {
	status := http.StatusInternalServerError
	switch err.(type) {
	case domain.InvalidGrantError:
		status = http.StatusBadRequest
	case domain.NotFoundError:
		status = http.StatusNotFound
	}
	return c.JSON(status, map[string]string{"error": err.Error()})
}
