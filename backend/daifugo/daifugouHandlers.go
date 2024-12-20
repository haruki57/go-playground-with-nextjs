package daifugo

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"math/rand/v2"
	"net/http"
	"slices"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type GameState string
const (
	WaitingForPlayers GameState = "WaitingForPlayers"
	PlayingCards GameState = "PlayingCards"
	GameEnded GameState = "GameEnded"
)

type PlayerRole string
const (
	Daifugo PlayerRole = "Daifugo"
	Fugo PlayerRole = "Fugo"
	Heimin PlayerRole = "Heimin"
	Hinmin PlayerRole = "Hinmin"
	Daihinmin PlayerRole = "Daihinmin"
)

type CardType string
const (
	Club CardType = "Club"
	Spade CardType = "Spade"
	Heart CardType = "Heart"
	Diamond CardType = "Diamond"
	Joker CardType = "Joker"
)

type Card struct {
	Number int `json:"number"`
	Value int `json:"value"`
	CardType CardType `json:"cardType"`
}
const JokerValue = 99

type CardStr = string // 3D, 13H, Joker etc...

type Player struct {
	Name string `json:"name"`
	Role PlayerRole `json:"role"`
	Cards []Card `json:"cards"`
}

type SubmitMode string
const (
	Normal SubmitMode = "Normal"
	ShibariMode SubmitMode = "ShibariMode"
	KakumeiMode SubmitMode = "KakumeiMode"
	KaidanMode SubmitMode = "KaidanMode"
)

type SpecialRule string
const (
Yagiri SpecialRule = "Yagiri"
KakumeiRule SpecialRule = "KakumeiRule"
ShibariRule SpecialRule = "ShibariRule"
Spade3Rule SpecialRule = "Spade3Rule"
//KaidanRule SpecialRule = "KaidanRule"
)
var StandardRule = map[SpecialRule]struct{}{
	Yagiri: {},
	KakumeiRule: {},
	ShibariRule: {},
	Spade3Rule: {},
}

type Result struct {
	GameNum int
	PlayersByRank []string
}

type Game struct {
	Players []*Player 
	GameState GameState 
	Turn int
	LastSubmittedTurn int
	SubmitModes map[SubmitMode]struct{}
	SpecialRules map[SpecialRule]struct{}
	PlayingCards []Card
	LastSubmittedNum int
	Trush []Card
	PlayersByRank []string
	PassCount int
	Results []Result
}

type DaifugoRoom struct {
	clients map[string]*websocket.Conn
	game *Game 
	mu sync.Mutex
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	rooms = make(map[string]*DaifugoRoom)
	mu    sync.Mutex
)

func createGameWithStandardRules() *Game {
	return &Game{
		Players: make([]*Player, 0),
		GameState: WaitingForPlayers,
		SpecialRules: maps.Clone(StandardRule),
		PlayingCards: make([]Card, 0),
		Turn: 0,
		LastSubmittedTurn: -1,
		PlayersByRank: make([]string, 0),
		Results: make([]Result, 0),
	}
}

func makeCard(num int, cardType CardType) Card {
	v := num
	if num <= 2 {
		v += 13
	}
	if cardType == Joker {
		v = JokerValue
	}
	return Card{num, v, cardType}
}

func makeDeck() []Card {
	total := 54 // 4 * 13 + 2
	ret := make([]Card, 0, total)
	ret = append(ret, makeCard(-1, Joker))
	ret = append(ret, makeCard(-2, Joker)) // set num to -2 to distinguish them
	for _, v := range []CardType{Club, Spade, Heart, Diamond} {
		for i := 1; i <= 13; i++ {
			ret = append(ret, makeCard(i, v))
		}	
	}
	rand.Shuffle(total, func(i, j int) {ret[i], ret[j] = ret[j], ret[i]})
	return ret
}

func (game *Game) startGame() error {
	if len(game.Players) <= 1 {
		return errors.New("num of players is not enough")
	}
	if len(game.Players) >= 7 {
		return errors.New("num of players is too many")
	}
	game.GameState = PlayingCards
	rand.Shuffle(len(game.Players), func(i, j int) {game.Players[i], game.Players[j] = game.Players[j], game.Players[i]})
	for _, player := range game.Players {
		player.Cards = make([]Card, 0)
	}
	for i, card := range makeDeck() {
		if i >= 4 {
			break
		}
		game.Players[i%len(game.Players)].Cards = append(game.Players[i%len(game.Players)].Cards, card)
	}
	if len(game.Results) >= 1 {
		previousResult := game.Results[len(game.Results)-1]
		for i, prevPlayer := range previousResult.PlayersByRank {
			for _, player := range game.Players {
				if prevPlayer == player.Name {
					player.Role = decideRole(i+1, len(game.Players))
					break
				}	
			}
		}
	} else {
		for _, player := range game.Players {
			player.Role = Heimin
		}
	}
	return nil
}

func decideRole(rank, totalPlayers int) PlayerRole {
	switch totalPlayers {
		case 2: 
		return []PlayerRole{Daifugo, Daihinmin}[rank-1]
		case 3: 
		return []PlayerRole{Daifugo, Heimin, Daihinmin}[rank-1]
		case 4: 
		return []PlayerRole{Daifugo, Fugo, Hinmin, Daihinmin}[rank-1]
		case 5: 
		return []PlayerRole{Daifugo, Fugo, Heimin, Hinmin, Daihinmin}[rank-1]
		case 6: 
		return []PlayerRole{Daifugo, Fugo, Heimin, Heimin, Hinmin, Daihinmin}[rank-1]
		default: 
		return Heimin
	}
}

func (game *Game) addPlayer(playerName string) error {
	for _, player := range game.Players {
		if player.Name == playerName {
			return errors.New("duplicated player name")
		}
	}
	game.Players = append(game.Players, &Player{Name: playerName})
	return nil
}

func (game *Game) removePlayer(playerName string) error {
	for i, player := range game.Players {
		if player.Name == playerName {
			game.Players = append(game.Players[:i], game.Players[i+1:]...)
			return nil
		}
	}
	return errors.New("cannot find player:" + playerName)
}

func (game *Game) getCurrentPlayer() *Player {
	if game.Turn < 0 {
		return nil
	}
	return game.Players[game.Turn]
}

func (game *Game) pass() {
	game.advanceTurn()
	game.PassCount++
	activePlayerNum := len(game.Players) - len(game.PlayersByRank)
	if game.Turn == game.LastSubmittedTurn || game.PassCount == activePlayerNum {
		game.discardPlayingCards()
	}
}
func (game *Game) getTopFieldCards() []Card { 
	return game.PlayingCards[len(game.PlayingCards)-game.LastSubmittedNum:]
}

func (game *Game) tryToSubmitCards(player *Player, submittingCards []Card) (isSubmitted bool, reason string) {
	currentPlayer := game.Players[game.Turn]
	if player.Name != currentPlayer.Name {
		return false, "not your turn"
	}

	// TODO validate that player.Cards contains submittingCards
	if canSubmit, reason := game.canSubmitCards(submittingCards); !canSubmit {
		return false, reason
	}
	game.LastSubmittedNum = len(submittingCards)
	game.PlayingCards = append(game.PlayingCards, submittingCards...)
	game.LastSubmittedTurn = game.Turn
	game.advanceTurn()	

	// Yagiri
	contains8 := false
	for _, card := range submittingCards {
		if card.Value == 8 {
			contains8 = true
			break
		}
	}
	spade3 := false
	lenPlayingCards := len(game.PlayingCards)
	if lenPlayingCards >= 2 && 
		game.PlayingCards[lenPlayingCards-2].CardType == Joker && 
		(game.PlayingCards[lenPlayingCards-1].Number == 3 && 
		game.PlayingCards[lenPlayingCards-1].CardType == Spade) {
		spade3 = true
	}

	// Nagasu
	if contains8 || spade3 {
		game.discardPlayingCards()
		game.Turn = game.LastSubmittedTurn
		game.LastSubmittedNum = 0
	}

	if len(submittingCards) >= 4 {
		game.flipKakumei();
	}

	player.removeCards(submittingCards)
	if len(player.Cards) == 0 {
		game.PlayersByRank = append(game.PlayersByRank, player.Name)
		if (len(game.PlayersByRank) == len(game.Players) - 1) {
			game.endGame()
		}
	}
	
	// TODO Shibari

	return true, "submitted"
}

func (game *Game) endGame() {
	game.GameState = GameEnded
	result := Result{}
	result.GameNum = len(game.Results) + 1
	result.PlayersByRank = game.PlayersByRank
	for _, player := range game.Players {
		found := true
		for _, playerName := range result.PlayersByRank {
			if player.Name == playerName {
				found = false
			}	
		}
		if found {
			result.PlayersByRank = append(result.PlayersByRank, player.Name)
		}
	}
	game.Results = append(game.Results, result)
}

func (game *Game) advanceTurn() {
	for ;; {
		game.Turn = (game.Turn + 1) % len(game.Players)	
		currentPlayer := game.getCurrentPlayer()
		if !slices.Contains(game.PlayersByRank, currentPlayer.Name) {
			break
		}
	}
}

func (player *Player) removeCards(cards []Card) {
	for _, card:= range cards {
		for i, playerCard := range player.Cards {
			if playerCard == card {
				player.Cards = player.Cards[:i+copy(player.Cards[i:], player.Cards[i+1:])]
			}
		}	
	}
}

func (game *Game) discardPlayingCards() {
	game.Trush = append(game.Trush, game.PlayingCards...)
	game.PlayingCards = make([]Card, 0)
	game.LastSubmittedNum = 0
	delete(game.SubmitModes, ShibariMode)
}

func (game *Game) flipKakumei() {
	if _, ok := game.SubmitModes[KakumeiMode]; ok {
		delete(game.SubmitModes, KakumeiMode)
	} else {
		game.SubmitModes[KakumeiMode] = struct{}{}
	}
}

func flipCardValue(cards []*Card) {
	for _, card := range cards {
		if card.CardType != Joker {
			card.Value *= -1
		}
	}
}

// TODO add special rule such as 縛り
func (game *Game) canSubmitCards(submittingCards []Card) (canSubmit bool, reason string) {
	if len(submittingCards) == 0 {
		return false, "no cards selected"
	}

	submitModes := game.SubmitModes
	specialRules := game.SpecialRules
	topFieldCards := game.getTopFieldCards()
	_, isKakumei := submitModes[KakumeiMode]
	if isKakumei {
		cards := make([]*Card, len(topFieldCards)+len(submittingCards))
		for i := 0; i < len(topFieldCards); i++ {
			cards[i] = &topFieldCards[i]
		}
		for i := 0; i < len(submittingCards); i++ {
			cards[i + len(topFieldCards)] = &submittingCards[i]
		}
		flipCardValue(cards)
		defer flipCardValue(cards)
	}

	isAllSameValue := true
	submitCardValue := JokerValue
	for i := 0; i < len(submittingCards); i++ {
		card1 := submittingCards[i]
		if card1.CardType == Joker {
			continue;
		} else {
			submitCardValue = card1.Value
		}
		for j := i+1; j < len(submittingCards); j++ {
			card2 := submittingCards[j]
			if card2.CardType == Joker {
				continue;
			}
			if card1.Value != card2.Value {
				isAllSameValue = false
			}	
		}
	}
	if !isAllSameValue {
		return false, "not all same value"
	}
	
	if len(topFieldCards) == 0 {
		return true, "no topFieldCards"
	}
	if len(topFieldCards) != len(submittingCards) {
		return false, "num of topFieldCards and submittingCards are different"
	}

	if _, isKaidan := submitModes[KaidanMode]; isKaidan {
		panic("KaidanMode is not implemented")
	}

	// _, isShibari := submitModes[ShibariMode]
	if _, isSpade3 := specialRules[Spade3Rule]; isSpade3 {
		if (len(topFieldCards) == 1 && len(submittingCards) == 1) &&
			(topFieldCards[0].CardType == Joker && submittingCards[0].Number == 3 && submittingCards[0].CardType == Spade) {
			return true, "spade 3 rule"
		}
	}

	
	currentValue := JokerValue
	for _, card := range topFieldCards {
		if currentValue > card.Value {
			currentValue = card.Value
		}
	}

	if currentValue < submitCardValue {
		return true, "submitted value is bigger"
	}
	return false, "no 'true' reason"
}

// getOrCreateRoom retrieves a room by name or creates a new one if it doesn't exist.
func getOrCreateRoom(roomName string) *DaifugoRoom {
	mu.Lock()
	defer mu.Unlock()

	room, exists := rooms[roomName]
	if !exists {
		room = &DaifugoRoom{
			clients: make(map[string]*websocket.Conn),
			game: createGameWithStandardRules(),
		}
		rooms[roomName] = room
	}
	return room
}

func (game *Game) addRule() {
	panic("TODO")
}

// WebSocketDaifugoHandler handles WebSocket connections for a specific room
func WebSocketDaifugoHandler(c *gin.Context) {
	roomName := c.Param("roomName")
	playerName := c.Param("playerName")
	room := getOrCreateRoom(roomName)
	room.game.addPlayer(playerName)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Add the client to the room
	room.mu.Lock()
	room.clients[playerName] = conn
	room.mu.Unlock()

	defer func() {
		room.mu.Lock()
		delete(room.clients, playerName)
		room.mu.Unlock()
	}()

	type AddPlayerDataResponse struct {
		PlayerNames []string `json:"playerNames"`
	}
	
	playerNames := make([]string, len(room.game.Players))
	for i, player := range room.game.Players {
		playerNames[i] = player.Name	
	}
	dataResponse, _ := json.Marshal(AddPlayerDataResponse{
		PlayerNames: playerNames,
	})
	responseObject := RawMessageResponse{
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

	// Listen for messages from the client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
		room.mu.Lock()
		handleWebsocketMessage(room, message)
		room.mu.Unlock()
		
		/*
		message = []byte(roomName + " " + (string(message)))
		for i, v := range roomName {
			message[i]=byte(v);
		}
		log.Printf("Room [%s] received: %s", roomName, message)

		// Broadcast the message to all clients in the room
		
		for client := range room.clients {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				client.Close()
				delete(room.clients, client)
			}
		}
		
		*/
	}
}

func DebugGetGameState(ctx *gin.Context) {
	roomName := ctx.Param("roomName")
	if rooms[roomName] == nil {
		ctx.JSON(http.StatusOK, nil)
	} else {
		ctx.JSON(http.StatusOK, rooms[roomName].game)
	}
}

func ListRoomsHandler(ctx *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	ret := make([]string, 0, len(rooms))
	for key, value := range rooms {
		fmt.Println(key)
		fmt.Println(value)
		ret = append(ret, key)
	}
	ctx.JSON(http.StatusOK, ret)
}

// /postMessage エンドポイント
func CreateRoomHandler(ctx *gin.Context) {
	roomName := ctx.Param("roomName")
	getOrCreateRoom(roomName)
	ctx.JSON(http.StatusOK, true)
}
