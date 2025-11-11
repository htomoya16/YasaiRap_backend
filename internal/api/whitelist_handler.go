package api

import (
	"backend/internal/service"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type WhitelistHandler struct {
	svc service.WhitelistService
}

func NewWhitelistHandler(s service.WhitelistService) *WhitelistHandler {
	return &WhitelistHandler{svc: s}
}

// Discord ID と VRC displayName を受け取り、
// VRChat APIで完全一致ユーザーを検索して whitelist_users に登録/更新する。
func (h *WhitelistHandler) RegisterDiscordVRC(c echo.Context) error {
	type RegisterDiscordVRCRequest struct {
		DiscordUserID  string `json:"discord_user_id"`
		VRCDisplayName string `json:"vrc_display_name"`
	}

	var r RegisterDiscordVRCRequest
	// 400
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid json: "+err.Error())
	}
	// 400
	if r.DiscordUserID == "" || r.VRCDisplayName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "discord_user_id and vrc_display_name are required")
	}

	created, err := h.svc.RegisterDiscordVRC(
		c.Request().Context(),
		r.DiscordUserID,
		r.VRCDisplayName,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidArgument):
			return echo.NewHTTPError(http.StatusBadRequest, "invalid argument")
		case errors.Is(err, service.ErrNoExactMatch):
			// VRChat Search All Users に完全一致が無かった
			return echo.NewHTTPError(http.StatusBadRequest, "no exact-matched vrchat user found for given display name")
		case errors.Is(err, service.ErrMultipleExactMatch):
			// 同じdisplayNameのユーザーが複数いて特定できない
			return echo.NewHTTPError(http.StatusBadRequest, "multiple vrchat users found with same display name")
		case errors.Is(err, service.ErrAlreadyExists):
			// その VRC userId は別のDiscordユーザーに既に紐づいている
			return echo.NewHTTPError(http.StatusConflict, "vrchat account already linked to another discord user")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}

	// 必要に応じて情報返したいならここで JSON 返す
	// （今は最低限フラグだけ）
	return c.JSON(status, map[string]any{
		"created": created,
	})
}

// 指定Discordユーザーの紐付け解除
func (h *WhitelistHandler) RemoveDiscordVRC(c echo.Context) error {
	type UnlinkDiscordVRCRequest struct {
		DiscordUserID string `json:"discord_user_id"`
	}

	var r UnlinkDiscordVRCRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid json: "+err.Error())
	}
	if r.DiscordUserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "discord_user_id is required")
	}

	if err := h.svc.RemoveDiscord(
		c.Request().Context(),
		r.DiscordUserID,
	); err != nil {
		if errors.Is(err, service.ErrInvalidArgument) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid argument")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
