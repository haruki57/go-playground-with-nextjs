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

type RawMessageResponse struct {
	Type string `json:"type"`
	Data json.RawMessage `json:"data"`
}

/*
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
	err := room.game.addPlayer(addPlayerDataRequest.PlayerName)
	if err != nil {
		log.Println(err)
		return
	}

	playerName := addPlayerDataRequest.PlayerName
	dataResponse, _ := json.Marshal(AddPlayerDataResponse{
		PlayerName: playerName,
	})
	responseObject := MessageResponse{
		Type: "ADD_PLAYER",
		Data: dataResponse,
	}
	response, _ := json.Marshal(responseObject)
	for playerName, client := range room.clients {
		if err := client.WriteMessage(websocket.TextMessage, response); err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			delete(room.clients, playerName)
		}
	}
}
*/
type RemovePlayerDataRequest struct {
	PlayerName string `json:"playerName"`
}

type RemovePlayerDataResponse struct {
	PlayerName string `json:"playerName"`
}

func handleRemovePlayer(room *DaifugoRoom, data json.RawMessage) {
	fmt.Println("handleRemovePlayer")
	var removePlayerDataRequest RemovePlayerDataRequest 
	json.Unmarshal(data, &removePlayerDataRequest)
	err := room.game.removePlayer(removePlayerDataRequest.PlayerName)
	if err != nil {
		log.Println(err)
		return
	}

	playerName := removePlayerDataRequest.PlayerName
	dataResponse, _ := json.Marshal(RemovePlayerDataResponse{
		PlayerName: playerName,
	})
	responseObject := RawMessageResponse{
		Type: "REMOVE_PLAYER",
		Data: dataResponse,
	}
	response, _ := json.Marshal(responseObject)
	for playerName, client := range room.clients {
		if err := client.WriteMessage(websocket.TextMessage, response); err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			delete(room.clients, playerName)
		}
	}
}

type GameStartRequest struct {
}

type GameStartResponse struct {
	/*    handCards: Card[];
    otherPlayerCards: Player[];*/
	HandCards []Card `json:"handCards"`
	Players []PublicPlayer `json:"players"`
}

type PublicPlayer struct{
	Name string `json:"name"`
	NumHandCards int `json:"numHandCards"`
}

func handleGameStart(room *DaifugoRoom) {
	fmt.Println("handleGameStart")
	game := room.game
	game.startGame()
	players := make([]PublicPlayer, len(game.Players))
	for i, player := range game.Players {
		players[i] = PublicPlayer{Name: player.Name, NumHandCards: len(player.Cards)}
	}
	fmt.Println(game)
	fmt.Println(game.Players)
	fmt.Println(players)
	for _, player := range game.Players {
		fmt.Println(player.Cards)	
	}
	for playerName, client := range room.clients {
		var handCards []Card
		for _, player := range game.Players {
			if player.Name == playerName {
				handCards = player.Cards
			}
		}
		gameStartResponse := GameStartResponse{
			HandCards: handCards,
			Players: players,
		}
		dataBytes, _ := json.Marshal(gameStartResponse)
		responseObject := RawMessageResponse{
			Type: "GAME_START",
			Data: dataBytes,
		}
		response, _ := json.Marshal(responseObject)
		if err := client.WriteMessage(websocket.TextMessage, response); err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			delete(room.clients, playerName)
		}
	}
}


type SubmitCardsRequest struct {
	/*    handCards: Card[];
		otherPlayerCards: Player[];*/
	PlayerName string `json:"playerName"`
	Cards []Card `json:"cards"`
}

type MessageResponse struct{
	Message string `json:"message"`
}
func handleSubmitCards(room *DaifugoRoom, data json.RawMessage) {
	fmt.Println("handleSubmitCards")
	var submitCardsRequest SubmitCardsRequest 
	json.Unmarshal(data, &submitCardsRequest)
	game := room.game
	canSubmit, reason := game.canSubmitCards(submitCardsRequest.Cards)
	fmt.Println(submitCardsRequest.Cards)
	fmt.Println(canSubmit)
	fmt.Println(reason)
	if (!canSubmit) {
		for playerName, client := range room.clients {
			fmt.Println(submitCardsRequest.PlayerName, playerName)
			if playerName != submitCardsRequest.PlayerName {
				continue
			}
			messageResponse := MessageResponse{"そのカードは出せません"}
			messageResponseBytes, _ := json.Marshal(messageResponse)
			responseObject := RawMessageResponse{
				Type: "MESSAGE",
				Data: messageResponseBytes,
			}
			response, _ := json.Marshal(responseObject)
			if err := client.WriteMessage(websocket.TextMessage, response); err != nil {
				log.Printf("WebSocket write error: %v", err)
				client.Close()
				delete(room.clients, playerName)
			}
			return
		}
		return
	}

	/*
	playerName := removePlayerDataRequest.PlayerName
	dataResponse, _ := json.Marshal(RemovePlayerDataResponse{
		PlayerName: playerName,
	})
	responseObject := RawMessageResponse{
		Type: "REMOVE_PLAYER",
		Data: dataResponse,
	}
	response, _ := json.Marshal(responseObject)
	for playerName, client := range room.clients {
		if err := client.WriteMessage(websocket.TextMessage, response); err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			delete(room.clients, playerName)
		}
	}
		*/
}

func handleWebsocketMessage(room *DaifugoRoom, rawMessage []byte) {
	fmt.Println("handleWebsocketMessage")
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
	//case "ADD_PLAYER":
//		handleAddPlayer(room, message.Data)
	case "REMOVE_PLAYER": 
		handleRemovePlayer(room, message.Data)
		case "GAME_START": 
		handleGameStart(room)
		case "SUBMIT_CARDS": 
		handleSubmitCards(room, message.Data)
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


