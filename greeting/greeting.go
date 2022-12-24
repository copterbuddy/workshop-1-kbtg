package greeting

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type handler struct{}

func (h *handler) Greeting(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func NewGreetingHandler() *handler {
	return &handler{}
}