package daifugo

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string      `json:"type"`
	PlayerName string `json:"playerName"`
	Data json.RawMessage `json:"data"`
}

// クライアントからのリクエスト
type SubmitCardData struct {
	Cards []string `json:"cards"`
}

type PassData struct{}

// サーバーからのレスポンス
type PlayerActionData struct {
	Player string   `json:"player"`
	Action string   `json:"action"`
	Cards  []string `json:"cards,omitempty"`
}

type GameUpdateData struct {
	CurrentTurn string       `json:"currentTurn"`
	LastPlayed  []string     `json:"lastPlayed"`
	Players     []PlayerInfo `json:"players"`
}

type PlayerInfo struct {
	Name          string `json:"name"`
	Role          string `json:"role"`
	RemainingCards int    `json:"remainingCards"`
}

type ErrorData struct {
	Message string `json:"message"`
}

var re = regexp.MustCompile(`^(\d+)([SHDC])$`)

func cardStrToCard(cardStr string) (Card, error) {
	cardStr = strings.ToUpper(cardStr)
	if cardStr == "JOKER" {
		return makeCard(-1, Joker), nil
	}
	matches := re.FindStringSubmatch(cardStr)
	if len(matches) != 3 {
		return Card{}, fmt.Errorf("invalid card format: %s", cardStr)
	}

	// 数字部分をintに変換
	number, err := strconv.Atoi(matches[1])
	if err != nil {
		return Card{}, fmt.Errorf("invalid number in card: %s", cardStr)
	}
	var cardType CardType
	switch matches[2] {
	case "S":
		cardType = Spade
	case "C":
		cardType = Club
	case "D":
		cardType = Diamond
	case "H":
		cardType = Heart
	}
	return makeCard(number, cardType), nil
}

func parseMessageTypeAndPlayerName(rawJsonBytes []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(rawJsonBytes, &msg)
	if err != nil {
		return Message{}, err
	}
	fmt.Println(msg.Data)
	return msg, nil
}

func handlePass(game *Game) {
	
}

type MessageResponse struct {
	Type string `json:"type"`
	Data json.RawMessage `json:"data"`
}

type AddPlayerDataRequest struct {
	PlayerName string `json:"playerName"`
}

type AddPlayerDataResponse struct {
	PlayerName string `json:"playerName"`
}

func handleAddPlayer(room *DaifugoRoom, data json.RawMessage) {
	fmt.Println("handleAddPlayer")
	var addPlayerDataRequest AddPlayerDataRequest 
	json.Unmarshal(data, &addPlayerDataRequest)
	room.game.addPlayer(addPlayerDataRequest.PlayerName)

	dataResponse, _ := json.Marshal(AddPlayerDataResponse{
		PlayerName: addPlayerDataRequest.PlayerName,
	})
	responseObject := MessageResponse{
		Type: "ADD_PLAYER",
		Data: dataResponse,
	}
	response, _ := json.Marshal(responseObject)
	for client := range room.clients {
		if err := client.WriteMessage(websocket.TextMessage, response); err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			delete(room.clients, client)
		}
	}
}

func handleWebsocketMessage(room *DaifugoRoom, rawMessage []byte) {
	fmt.Println("handleWebwocketMessage")
	message, err := parseMessageTypeAndPlayerName(rawMessage)
	fmt.Println(message)
	if err != nil {
		log.Printf("MessageType parse error: %v", err)
		return
	}
	//game := room.game
	//currentPlayer := game.Players[game.Turn]

	// send "not your turn" message
	//if game.GameState != PlayingCards && currentPlayer.Name != message.PlayerName {
		//return
	//}
	
	switch message.Type {
	case "ADD_PLAYER":
		handleAddPlayer(room, message.Data)
	case "PASS":
		//game.pass(currentPlayer)
		/*
		currentTurn := game.Turn
		type Hoge struct {
			Turn int `json:"turn"`
			Player Player `json:"player"`
		}
*/
	}

}