// main.go
package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	// Ginのインスタンス作成
	router := gin.Default()

	// /events エンドポイントを登録
	router.GET("/events", SSEHandler)

	// サーバーを起動
	router.Run(":8080")
}
