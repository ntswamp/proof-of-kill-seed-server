package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"app/src/handler"
	lib_log "app/src/lib/log"

	"github.com/gin-gonic/gin"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	engine := gin.Default()
	handler.RouteApi(engine)

	addr := ":9090"

	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			lib_log.Error("listen: %v", err)
		}

	}()
	// サーバが停止したら10秒待ってからシャットダウン.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	lib_log.Info("shutdown server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		lib_log.Error("server shutdown:", err)
		return
	}
	lib_log.Info("server exited")
}
