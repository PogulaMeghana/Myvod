package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func (uh *UsersHandler) Healthz(c echo.Context) error {
	return c.JSON(http.StatusOK, "Welcome to Users Service")
}
