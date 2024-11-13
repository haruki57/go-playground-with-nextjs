package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func sseHandler(w http.ResponseWriter, r *http.Request) {
	// ヘッダーの設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// SSEイベントを送信するループ
	for {
		fmt.Fprintf(w, "data: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
		// バッファをフラッシュしてクライアントにデータを送信
		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}

		// 1秒間スリープ
		time.Sleep(1 * time.Second)
	}
}

func main() {
	http.HandleFunc("/events", sseHandler)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
