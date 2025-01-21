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
- Wrap around?

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

type HighScoreMessage struct {
	Name  string
	Score int
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
	includeStuff("client.html")
	game.InitGame()
	go gameLoop()

	http.HandleFunc("/join", joinGame)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "out/client.html")
	})

	http.HandleFunc("/get-status", status)
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "status.html")
	})

	err := http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", nil)
	if err != nil {
		println(":(")
		println(err.Error())
	}
}

type ClientInfo struct {
	Name          string
	RemoteAddress string
	LocalAddress  string
}

func status(w http.ResponseWriter, r *http.Request) {
	statusMessage := make([]ClientInfo, len(clients))
	for i := 0; i < len(clients); i++ {
		if clients[i] != nil {
			statusMessage[i].Name = clients[i].player.Name
			statusMessage[i].RemoteAddress = clients[i].connection.RemoteAddr().String()
			statusMessage[i].LocalAddress = clients[i].connection.LocalAddr().String()
		}

	}
	jsonMessage, _ := json.Marshal(statusMessage)
	w.Write(jsonMessage)
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
			broadcastPlayerList(&clients)
			broadcastByteMessageToActiveClients(&clients, &HighScoreMessage{
				Name:  gameState.state.HighScore.Name,
				Score: gameState.state.HighScore.Score,
			})
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
	gameJoinRequest.Validate()

	// TODO: Validate the request.
	// Check that color is hex and name is not too long

	fmt.Printf("Name: %s\n", gameJoinRequest.SnakeName)
	fmt.Printf("Color: %s\n", gameJoinRequest.SnakeColor)

	client, err := createClient(conn, gameJoinRequest.SnakeColor, gameJoinRequest.SnakeName)
	if err != nil {
		println(err.Error())
		sendMessage(conn, &TextInfoMessage{"Cannot connect. Too many players"})

		conn.Close()
		return
	}

	joinMessage := gameJoinRequest.SnakeName + " connected"
	broadcastByteMessageToActiveClients(&clients, &TextInfoMessage{joinMessage})

	gameSetup := GameSetupMessage{LevelWidth: game.LevelWidth, LevelHeight: game.LevelHeight}
	sendMessage(client.connection, &gameSetup)

	broadcastPlayerList(&clients)
	go inputLoop(client)
	println("Game started")
}

func broadcastPlayerList(clients *[]*Client) {
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
	//	var message, _ = json.Marshal(gameState.state)
	for i := 0; i < len(clients); i++ {
		if clients[i] == nil {
			continue
		}
		if !clients[i].connected {
			err := clients[i].connection.Close()
			if err != nil {
				println("Failed to close connection")
				clients[i] = nil
				continue
			}
			broadcastByteMessageToActiveClients(&clients, &TextInfoMessage{clients[i].player.Name + " disconnected"})
			clients[i] = nil
			continue
		}
		var err = sendMessage(clients[i].connection, &gameState)
		//var err = sendMessageToClient(clients[i].connection, GameStateUpdate, message)
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
