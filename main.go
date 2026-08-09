// Package main 是程式進入點。
//
// Go 的規則:可執行程式的 package 必須叫 main,且要有一個 func main()。
// 這相當於 Java 的 public static void main(String[] args)。
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// 匯入我們自己的 package。路徑 = go.mod 的 module 名稱 + 資料夾路徑。
	"taskapi/internal/task"
)

func main() {
	// 1. 組裝依賴(手動 DI,不需要 Spring)。
	store := task.NewMemoryStore()
	handler := task.NewHandler(store)

	// 2. 設定 HTTP 伺服器。設 timeout 是正式環境的好習慣,避免慢速連線拖垮服務。
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler.Routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 3. 在一個 goroutine 裡啟動伺服器。
	//    「go」關鍵字就開一條 goroutine——比 Java 的 new Thread() 輕量太多,
	//    一個程式可以輕鬆同時跑成千上萬條。ListenAndServe 會阻塞,所以放進 goroutine。
	go func() {
		log.Println("listening on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 4. 優雅關閉(graceful shutdown):
	//    等待作業系統送來的中斷訊號(Ctrl+C / kill),收到後給正在處理的請求
	//    最多 10 秒收尾時間再關閉。這展示了 channel 與 context 的實務用法。
	quit := make(chan os.Signal, 1)             // 建立一個帶緩衝的 channel
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit                                       // 阻塞,直到 channel 收到訊號
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
