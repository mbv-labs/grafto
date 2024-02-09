package middleware

import (
	"net/http"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Middleware struct {
	services services.Services
}

func NewMiddleware(services services.Services) Middleware {
	return Middleware{services}
}

type AuthContext struct {
	echo.Context
	userID          uuid.UUID
	isAuthenticated bool
}

func (a *AuthContext) GetID() uuid.UUID {
	return a.userID
}
func (a *AuthContext) GetAuthStatus() bool {
	return a.isAuthenticated
}

type AdminContext struct {
	echo.Context
	isAdmin bool
}

func (a *AdminContext) GetAdminStatus() bool {
	return a.isAdmin
}

func (m *Middleware) AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authenticated, userID, err := m.services.IsAuthenticated(c.Request())
		if err != nil {
			telemetry.Logger.Error("could not get authenticated status", "error", err)
			return c.Redirect(http.StatusPermanentRedirect, "/500")
		}

		if authenticated {
			ctx := &AuthContext{c, userID, true}
			return next(ctx)
		} else {
			return c.Redirect(http.StatusPermanentRedirect, "/login")
		}
	}
}

func (m *Middleware) AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		isAdmin, err := m.services.IsAdmin(c.Request())
		if err != nil {
			return c.Redirect(http.StatusPermanentRedirect, "/500")
		}

		ctx := &AdminContext{c, isAdmin}
		return next(ctx)
	}
}
