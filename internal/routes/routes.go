package routes

import (
	"backend/internal/handler"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(
	// 引数
	e *echo.Echo,
	healthHandler *handler.HealthHandler) {

	api := e.Group("/api")

	// ヘルスチェック
	api.GET("/livez", healthHandler.Livez)
	api.GET("/readyz", healthHandler.Readyz)
	api.GET("/healthz", healthHandler.Healthz)
}
