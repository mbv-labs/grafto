package views

import (
	"github.com/MBvisti/grafto/views/internal/layouts"
	"github.com/MBvisti/grafto/views/internal/templates"
	"github.com/labstack/echo/v4"
)

func Dashboard(ctx echo.Context) error {
	return layouts.Dashboard(templates.DashboardIndex()).Render(extractRenderDeps(ctx))
}
