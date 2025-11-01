package main

import (
	"backend/database"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// DB接続
	db, err := database.NewConnection()
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer db.Close()

	// Echo インスタンスを作成
	e := echo.New()

	// ミドルウェア
	// 起動時のASCIIバナーを消す
	e.HideBanner = true
	// /users/ のような末尾スラッシュをリクエスト内で剥がして /users に書き換える
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(
		middleware.Recover(),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.CORS(),
	)

	// ポート設定
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ---- server with timeouts ----
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      e,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ---- start & wait for signal ----
	errCh := make(chan error, 1)
	go func() { errCh <- e.StartServer(srv) }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		e.Logger.Info("Server is shutting down...")
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal(err)
		}
	}

	// ---- graceful shutdown ----
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		e.Logger.Error("graceful shutdown failed, forcing close:", err)
		if cerr := e.Close(); cerr != nil {
			e.Logger.Error(cerr)
		}
	}

	// DBはここで閉じる（全リクエスト完了後）
	if derr := db.Close(); derr != nil {
		e.Logger.Error("db close:", derr)
	}

	e.Logger.Info("Server stopped")

}
