package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go_project/game"
	"go_project/requests"
	"net/http"
	"time"
)

/*
Bugs

TODO
-

479 bytes per meddelande med json

Pickup ideas
- Faster speed
- Slower speed
- Invisible
- Bombs


Anti cheat
Send some hash to show code is not modified. What about replay attacks?
*/

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type GameStateMessageWrapper struct {
	state game.GameStateMessage
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
	Score int
}

type GameSetupMessage struct {
	LevelWidth  int
	LevelHeight int
}

type TextInfoMessage struct {
	Text string
}

var clients = make([]*Client, 20)

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
		gameState := GameStateMessageWrapper{state: game.Tick(activeClients)}

		broadcastGameState(gameState)
		if gameState.state.ScoreChanged {
			broadcastPlayerlist(&clients)
		}

		time.Sleep(80 * time.Millisecond)
	}
}

/*
Note: CheckOrigin in Upgrader allows all connections.
In a production environment, make sure to validate the origin to avoid Cross-Site WebSocket Hijacking.
*/
func joinGame(responseWriter http.ResponseWriter, request *http.Request) {

	fmt.Printf("%s connected\n", request.Host)

	conn, upgradeError := upgrader.Upgrade(responseWriter, request, nil)
	if upgradeError != nil {
		println("Failed to upgrade")
		println(upgradeError.Error())
		return
	}
	_, msg, msgError := conn.ReadMessage()

	if msgError != nil {
		println("Error when reading join message")
		return
	}

	gameJoinRequest := requests.GameJoinRequest{}
	json.Unmarshal(msg, &gameJoinRequest)

	fmt.Printf("Name: %s\n", gameJoinRequest.SnakeName)
	fmt.Printf("Color: %s\n", gameJoinRequest.SnakeColor)

	client, err := createClient(conn, gameJoinRequest.SnakeColor, gameJoinRequest.SnakeName)
	if err != nil {
		println(err.Error())
		sendMessage(conn, &TextInfoMessage{"Cannot connect. Too many players"})

		conn.Close()
		return
	}
	sendMessage(conn, &TextInfoMessage{"Connected"})

	gameSetup := GameSetupMessage{LevelWidth: game.LevelWidth, LevelHeight: game.LevelHeight}
	sendMessage(client.connection, &gameSetup)

	broadcastPlayerlist(&clients)
	go inputLoop(client)
	println("Game started")
}

func broadcastPlayerlist(clients *[]*Client) {
	playerListUpdate := PlayerListUpdateMessage{make([]PlayerListEntry, 0)}

	for i := 0; i < len(*clients); i++ {
		client := (*clients)[i]
		if client == nil || !client.connected {
			continue
		}
		playerListUpdate.Entries = append(playerListUpdate.Entries, PlayerListEntry{
			Name:  client.player.Name,
			Color: client.player.SnakeColor,
			Score: client.player.TailLength,
		})
	}
	broadcastByteMessageToActiveClients(clients, &playerListUpdate)
}

func createClient(connection *websocket.Conn, snakeColor string, name string) (*Client, error) {
	index, err := getFirstFreeSlotIndex(clients)
	if err != nil {
		return nil, err
	}
	player := game.CreatePlayer(name, snakeColor)

	client := Client{
		connection: connection,
		connected:  true,
		player:     player,
	}

	clients[index] = &client
	return &client, nil
}

func getFirstFreeSlotIndex(clients []*Client) (int, error) {
	for i := 0; i < len(clients); i++ {
		if clients[i] == nil || clients[i].connected == false {
			return i, nil
		}
	}
	return -1, &ServerFullError{}
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
		game.HandleInput(&c.player, input)
	}
}

func broadcastGameState(gameState GameStateMessageWrapper) {
	if len(gameState.state.Players) > 0 {
		gameState.state.Players[0].Tail = getCorners(gameState.state.Players[0].Tail)
	}
	var message, _ = json.Marshal(gameState.state)
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

func broadcastByteMessageToActiveClients(clients *[]*Client, message GameMessage) {
	for i := 0; i < len(*clients); i++ {
		client := (*clients)[i]
		if client == nil || !client.connected {
			continue
		}

		sendMessage(client.connection, message)
	}
}
