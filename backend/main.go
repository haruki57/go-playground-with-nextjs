// main.go
package main

import (
	"go-playground/daifugo"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CORSヘッダーの設定
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// プリフライトリクエストの対応
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}


func main() {
	// Ginのインスタンス作成
	router := gin.Default()

	// CORSミドルウェアの使用
	router.Use(CORSMiddleware())

	router.GET("/hello", func(ctx *gin.Context) {	ctx.JSON(http.StatusOK, 123)})

	// /events エンドポイントを登録
	router.GET("/events", SSEHandler)
	router.GET("/eventsJson", EventsJsonHandler)
	
	router.POST("/postMessage", PostMessageHandler)
	router.GET("/streamMessage", StreamMessageHandler)

	// WebSocket endpoint for rooms
	router.GET("/ws/:roomName", WebSocketHandler)

	// daifugo
	router.GET("/daifugo/debug/:roomName", daifugo.DebugGetGameState)
	router.GET("/daifugo/ws/:roomName", daifugo.WebSocketDaifugoHandler)

	// サーバーを起動
	router.Run(":8080")
}
