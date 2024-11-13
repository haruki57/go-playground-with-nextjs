// main.go
package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	// Ginのインスタンス作成
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	})
	// /events エンドポイントを登録
	router.GET("/events", SSEHandler)
	router.GET("/eventsJson", EventsJsonHandler)

	// サーバーを起動
	router.Run(":8080")
}
