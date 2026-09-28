package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/maneeshaindrachapa/eris/internal/domain"
)

func NewServer(authSvc domain.AuthorizeService, tokenSvc domain.TokenService, adminSvc domain.AdminService, frontendOrigin string) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontendOrigin},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	ah := &AuthorizeHandler{svc: authSvc}
	th := &TokenHandler{svc: tokenSvc}
	admin := &AdminHandler{svc: adminSvc}

	e.GET("/authorize", ah.Start)
	e.POST("/authorize/login", ah.Login)
	e.POST("/authorize/consent", ah.Consent)
	e.POST("/token", th.Token)
	e.GET("/admin/clients", admin.ListClients)
	e.POST("/admin/clients", admin.CreateClient)
	e.PUT("/admin/clients/:id", admin.UpdateClient)
	e.GET("/healthz", func(c echo.Context) error { return c.String(200, "ok") })

	return e
}
