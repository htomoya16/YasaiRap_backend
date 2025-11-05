package main

import (
	"backend/database"
	"backend/internal/handler"
	"backend/internal/repository"
	"backend/internal/routes"
	"backend/internal/service"
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

	// DI
	healthRepo := repository.NewHealthRepository(db)
	healthSevice := service.NewHealthService(healthRepo)
	healthHandler := handler.NewHealthHandler(healthSevice)

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

	// ルート設定
	routes.SetupRoutes(e, healthHandler)

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

	// ---- 起動ウォームアップ：依存OKならready ON ----
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		// まずflagをONにしてから疎通を見る。
		healthSevice.MarkReady()
		if !healthSevice.Ready(ctx) {
			// 依存がまだならflagをOFF
			healthSevice.MarkNotReady()
		}
	}()

	// ---- server start & wait for signal ----
	// サーバ起動結果（エラー）を受け取るためのチャネルを用意する（バッファ1で送信ブロックを避ける）
	errCh := make(chan error, 1)
	// サーバを別ゴルーチンで起動する
	go func() { errCh <- e.StartServer(srv) }()

	// OSシグナル（Ctrl+C の SIGINT と SIGTERM）を受けると自動で Done になるコンテキストを作る
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 「シグナルでの終了要求」か「サーバ起動側のエラー」のどちらが先かを競合待ちする
	select {
	case <-ctx.Done():
		// シグナルを受けたのでシャットダウンへ進む。 ログを出す。
		e.Logger.Info("Server is shutting down...")
	case err := <-errCh:
		// サーバ起動側が先に戻った（起動失敗 or 正常終了）
		if !errors.Is(err, http.ErrServerClosed) {
			// それ以外はポート競合などの致命的な起動失敗とみなして落とす
			e.Logger.Fatal(err)
		}
	}

	// ---- graceful shutdown ----
	// まずreadyを落としてロードバランサから外れる（ドレイン）
	healthSevice.MarkNotReady()
	// 猶予時間を設定（ここでは10秒）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 新規受付を止める
	if err := e.Shutdown(shutdownCtx); err != nil {
		// 猶予内に閉じられない等で失敗した場合はログに残し
		e.Logger.Error("graceful shutdown failed, forcing close:", err)
		// 最終手段として強制クローズ（未完リクエストはエラーになる前提）
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
