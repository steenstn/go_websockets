package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"go_project/game"
	"go_project/requests"
	"net/http"
	"time"
)

/*
Bugs

TODO
- Remove clients when connection is dropped. Max number on init?
-

479 bytes per meddelande med json

Pickup ideas
- Faster speed
- Slower speed
- Invisible
- Bombs


Anti cheat
Send some hash to show code is not modified
*/

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	connection *websocket.Conn
	connected  bool
	player     game.Player
}

type PlayerListUpdateMessage struct {
	Entries []PlayerListEntry
}

type PlayerListEntry struct {
	Name  string
	Color string
}

type GameSetupMessage struct {
	LevelWidth  int
	LevelHeight int
}

var clients = make([]*Client, 10)

func main() {
	game.InitGame()
	go gameLoop()

	http.HandleFunc("/join", joinGame)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "client.html")
	})
	http.ListenAndServe(":8080", nil)
}

// Do this with channels instead?
func gameLoop() {
	for {
		activeClients := make([]*game.Player, 0)
		for i := 0; i < len(clients); i++ {
			if clients[i] != nil && clients[i].connected {
				activeClients = append(activeClients, &clients[i].player)
			}
		}
		gameState := game.Tick(activeClients)

		broadcastGameState(gameState)

		time.Sleep(80 * time.Millisecond)
	}
}

/*
Note: CheckOrigin in Upgrader allows all connections.
In a production environment, make sure to validate the origin to avoid Cross-Site WebSocket Hijacking.
*/
func joinGame(responseWriter http.ResponseWriter, request *http.Request) {

	conn, _ := upgrader.Upgrade(responseWriter, request, nil)
	_, msg, msgError := conn.ReadMessage()

	if msgError != nil {
		println("Error when reading join message")
		return
	}

	gameJoinRequest := requests.GameJoinRequest{}
	json.Unmarshal(msg, &gameJoinRequest)

	client := createClient(conn, gameJoinRequest.SnakeColor, gameJoinRequest.SnakeName)
	//clients = append(clients, client)

	gameSetup := GameSetupMessage{LevelWidth: game.LevelWidth, LevelHeight: game.LevelHeight}
	outgoingMessage, _ := json.Marshal(gameSetup)
	sendMessageToClient(client.connection, GameSetup, outgoingMessage)

	playerListUpdate := PlayerListUpdateMessage{make([]PlayerListEntry, 0)}

	for i := 0; i < len(clients); i++ {
		if clients[i] == nil || !clients[i].connected {
			continue
		}
		playerListUpdate.Entries = append(playerListUpdate.Entries, PlayerListEntry{
			Name:  clients[i].player.Name,
			Color: clients[i].player.SnakeColor,
		})
	}
	broadcastMessageToActiveClients(&clients, PlayerListUpdate, playerListUpdate)

	go inputLoop(client)
	println("Game started")
}

func createClient(connection *websocket.Conn, snakeColor string, name string) *Client {

	index := getFirstFreeSlotIndex(clients)

	player := game.CreatePlayer(name, snakeColor)

	client := Client{
		connection: connection,
		connected:  true,
		player:     player,
	}

	clients[index] = &client
	return &client
}

func getFirstFreeSlotIndex(clients []*Client) int {
	for i := 0; i < len(clients); i++ {
		if clients[i] == nil || clients[i].connected == false {
			return i
		}
	}
	return -1
}

func inputLoop(c *Client) {
	println("Starting input loop")

	for {
		if c.connected == false {
			break
		}
		_, msg, err := c.connection.ReadMessage()
		if err != nil {
			println("Input reading failed, player dropped")
			c.connected = false
			//c.connection.Close()
			break
		}

		// TODO More sanitation
		if len(msg) > 100 {
			println("Too long message, not processing")
			continue
		}
		//message := string(msg)
		//fmt.Printf("%s sent: %s\n", c.connection.RemoteAddr(), message)

		var input = string(msg)
		game.SetWantedDirection(&c.player, input)

	}
}

func broadcastGameState(gameState game.GameStateMessage) {
	var message, _ = json.Marshal(gameState)
	for i := 0; i < len(clients); i++ {
		if clients[i] == nil {
			continue
		}
		if !clients[i].connected {
			clients[i].connection.Close()
			continue
		}
		var err = sendMessageToClient(clients[i].connection, GameStateUpdate, message)
		if err != nil {
			closeError := clients[i].connection.Close()
			clients[i].connected = false
			if closeError != nil {
				println("Failed to close connection")
			}
		}
	}
}

func broadcastMessageToActiveClients(clients *[]*Client, messageType MessageType, message any) {
	var jsonMessage, err = json.Marshal(message)
	if err != nil {
		println("Failed to marshall message")
	}

	for i := 0; i < len(*clients); i++ {
		client := (*clients)[i]
		if client == nil || !client.connected {
			continue
		}

		var err = sendMessageToClient(client.connection, messageType, jsonMessage)
		if err != nil {
			println("Failed to send message")
		}
	}
}
