package api

import (
	"github.com/labstack/echo/v4"
)

func SetupRoutes(
	// 引数
	e *echo.Echo,
	healthHandler *HealthHandler,
	whitelistHandler *WhitelistHandler) {

	api := e.Group("/api")

	// ヘルスチェック
	api.GET("/livez", healthHandler.Livez)
	api.GET("/readyz", healthHandler.Readyz)
	api.GET("/healthz", healthHandler.Healthz)

	// Discordホワイトリスト管理用
	discord := api.Group("/discord")
	// 登録/更新
	discord.POST("/whitelist/register", whitelistHandler.RegisterDiscordVRC)
	// 削除
	discord.POST("/whitelist/remove", whitelistHandler.RemoveDiscordVRC)
}
