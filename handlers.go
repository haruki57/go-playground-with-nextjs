package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SSEHandler(c *gin.Context) {
	// ヘッダーの設定
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// SSEイベントを送信するループ
	for {
		// 現在時刻を送信
		fmt.Fprintf(c.Writer, "data: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

		// バッファをフラッシュしてクライアントにデータを送信
		flusher, ok := c.Writer.(http.Flusher)
		if ok {
			flusher.Flush()
		}

		// 1秒間スリープ
		time.Sleep(1 * time.Second)
	}
}

// JSON形式でのイベント送信 from eventsJson.html
func EventsJsonHandler(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")

	for {
		data := map[string]string{
			"time": time.Now().Format("2006-01-02 15:04:05"),
			"msg":  "This is a JSON formatted event",
		}
		c.JSON(200, data)

		// バッファをフラッシュしてクライアントにデータを送信
		flusher, ok := c.Writer.(http.Flusher)
		if ok {
			flusher.Flush()
		}

		// 1秒間スリープ
		time.Sleep(1 * time.Second)
	}
}