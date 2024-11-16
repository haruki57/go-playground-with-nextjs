package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

	for i := 0; i < 3; i++ {
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

/* 
Implement by Mutex

// グローバル変数とロック
var message string
var mu sync.Mutex

// /postMessage エンドポイント
func PostMessageHandler(c *gin.Context) {
	var requestBody struct {
		Text string `json:"text"`
	}

	// JSONデータのバインド
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// メッセージを大文字に変換して保存
	mu.Lock()
	message = strings.ToUpper(requestBody.Text)
	mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": "Message received"})
}

// /streamMessage エンドポイント (SSE)
func StreamMessageHandler(c *gin.Context) {
	// クライアントにSSEを使うためのヘッダー設定
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// メッセージのロック
	mu.Lock()
	msg := message
	mu.Unlock()

	// メッセージをSSEとして送信
	if msg != "" {
		c.SSEvent("message", msg)
	}

	// フラッシュでクライアントにデータを送信
	flusher, ok := c.Writer.(http.Flusher)
	if ok {
		flusher.Flush()
	}

	// 更新頻度を設定（1秒間隔で送信）
	//time.Sleep(1 * time.Second)
}
*/


// メッセージ用のchannel
var messageChannel = make(chan string)

// /postMessage エンドポイント
func PostMessageHandler(c *gin.Context) {
	var requestBody struct {
		Text string `json:"text"`
	}

	// JSONデータのバインド
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// メッセージを大文字に変換してchannelに送信
	upperText := strings.ToUpper(requestBody.Text)
	messageChannel <- upperText
	c.JSON(http.StatusOK, gin.H{"status": "Message received"})
}

// /streamMessage エンドポイント (SSE)
func StreamMessageHandler(c *gin.Context) {
	// クライアントにSSEを使うためのヘッダー設定
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// necessary to end connection when tab is closed
	body := c.Request.Body
	defer body.Close()
	// channelからメッセージを受け取り、SSEで送信
	for {

		select {
		case msg := <-messageChannel:
			// メッセージをSSEイベントとして送信
			c.SSEvent("message", msg)
			
			// フラッシュでクライアントにデータを送信
			flusher, ok := c.Writer.(http.Flusher)
			if ok {
				flusher.Flush()
			}
		case <-time.After(1 * time.Second):
			// 定期的に接続を維持するための心拍イベント
			c.SSEvent("heartbeat", "keep-alive")
			flusher, ok := c.Writer.(http.Flusher)
			if ok {
				flusher.Flush()
			}
		case <-c.Request.Context().Done():
			// リクエストコンテキストがキャンセルされた場合（クライアントの接続が切れたとき）
			fmt.Println("Done")
			return
		}
	}
}

// websocket

type Room struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	rooms = make(map[string]*Room)
	mu    sync.Mutex
)

// getOrCreateRoom retrieves a room by name or creates a new one if it doesn't exist.
func getOrCreateRoom(roomName string) *Room {
	mu.Lock()
	defer mu.Unlock()

	room, exists := rooms[roomName]
	if !exists {
		room = &Room{
			clients: make(map[*websocket.Conn]bool),
		}
		rooms[roomName] = room
	}
	return room
}

// WebSocketHandler handles WebSocket connections for a specific room
func WebSocketHandler(c *gin.Context) {
	roomName := c.Param("roomName")
	room := getOrCreateRoom(roomName)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Add the client to the room
	room.mu.Lock()
	room.clients[conn] = true
	room.mu.Unlock()

	defer func() {
		room.mu.Lock()
		delete(room.clients, conn)
		room.mu.Unlock()
	}()

	// Listen for messages from the client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
		message = []byte(roomName + " " + (string(message)))
		for i, v := range roomName {
			message[i]=byte(v);
		}
		log.Printf("Room [%s] received: %s", roomName, message)

		// Broadcast the message to all clients in the room
		room.mu.Lock()
		for client := range room.clients {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				client.Close()
				delete(room.clients, client)
			}
		}
		room.mu.Unlock()
	}
}