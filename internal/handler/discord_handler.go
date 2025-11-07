package handler

import (
	"backend/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type DiscordHandler struct {
	svc service.DiscordService
}

func NewDiscordHandler(s service.DiscordService) *DiscordHandler {
	return &DiscordHandler{svc: s}
}

func (h *DiscordHandler) AddWhitelist(c echo.Context) error {
	type req struct{ UserID, Note string }
	var r req
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if r.UserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "userID required")
	}
	if err := h.svc.AddWhitelist(c.Request().Context(), r.UserID, r.Note); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusCreated)
}

func (h *DiscordHandler) RemoveWhitelist(c echo.Context) error {
	type req struct{ UserID string }
	var r req
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if r.UserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "userID required")
	}
	if err := h.svc.RemoveWhitelist(c.Request().Context(), r.UserID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
