package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

func NewServer(authSvc domain.AuthorizeService, tokenSvc domain.TokenService) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger())

	ah := &AuthorizeHandler{svc: authSvc}
	th := &TokenHandler{svc: tokenSvc}

	e.GET("/authorize", ah.Start)
	e.POST("/authorize/login", ah.Login)
	e.POST("/authorize/consent", ah.Consent)
	e.POST("/token", th.Token)
	e.GET("/healthz", func(c echo.Context) error { return c.String(200, "ok") })

	return e
}
