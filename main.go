package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go_project/requests"
	"math/rand"
	"net/http"
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
	direction       Direction
	wantedDirection Direction
	snake           []TailSegment
	connection      *websocket.Conn
	alive           bool
	tailLength      int
	snakeColor      string
	name            string
}

func toTailMessage(tailSegment []TailSegment, tailLength int) []TailMessage {
	result := make([]TailMessage, tailLength)
	for i := 0; i < tailLength; i++ {
		result[i].X = tailSegment[i].x
		result[i].Y = tailSegment[i].y
	}
	return result
}

type PlayerMessage struct {
	X         int
	Y         int
	Direction Direction
	Color     string
	Tail      []TailMessage
}

type TailMessage struct {
	X int
	Y int
}

type PickupMessage struct {
	X    int
	Y    int
	Type int
}

type GameStateMessage struct {
	Players []PlayerMessage
	Pickups []PickupMessage
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

var clients = make([]*Client, 0)
var pickups = make([]Pickup, 5)

func main() {
	initGame()
	gameRunning = true
	go gameLoop()

	http.HandleFunc("/join", joinGame)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "client.html")
	})
	http.ListenAndServe(":8080", nil)
}

func joinGame(responseWriter http.ResponseWriter, request *http.Request) {
	if gameRunning == false {
		println("No game running")
		return
	}
	conn, _ := upgrader.Upgrade(responseWriter, request, nil)
	_, msg, msgError := conn.ReadMessage()

	if msgError != nil {
		println("Error when reading join message")
		return
	}

	gameJoinRequest := requests.GameJoinRequest{}
	json.Unmarshal(msg, &gameJoinRequest)

	client := createClient(conn, gameJoinRequest.SnakeColor, gameJoinRequest.SnakeName)
	clients = append(clients, client)

	gameSetup := GameSetupMessage{LevelWidth: levelWidth, LevelHeight: levelHeight}
	outgoingMessage, _ := json.Marshal(gameSetup)
	sendMessageToClient(client.connection, GameSetup, outgoingMessage)

	playerListUpdate := PlayerListUpdateMessage{make([]PlayerListEntry, 0)}

	for i := 0; i < len(clients); i++ {
		playerListUpdate.Entries = append(playerListUpdate.Entries, PlayerListEntry{
			Name:  clients[i].name,
			Color: clients[i].snakeColor,
		})
	}
	broadcastMessageToActiveClients(&clients, PlayerListUpdate, playerListUpdate)

	go inputLoop(client)
	println("Game started")

}

func createClient(connection *websocket.Conn, snakeColor string, name string) *Client {
	client := Client{
		direction:       down,
		wantedDirection: down,
		connection:      connection,
		alive:           true,
		snake:           make([]TailSegment, 100),
		tailLength:      5,
		snakeColor:      snakeColor,
		name:            name,
	}
	client.snake[0].x = (10 + 10*len(clients)) % levelWidth
	client.snake[0].y = 10
	return &client
}

func initGame() {
	println("Initiating game")

	for i := 0; i < len(pickups); i++ {
		pickups[i].pickupType = 0
		pickups[i].x = rand.Intn(2 + levelWidth - 4)
		pickups[i].y = rand.Intn(2 + levelHeight - 4)
	}
}

func inputLoop(c *Client) {
	println("Starting input loop")
	for {
		if c.alive == false {
			break
		}
		_, msg, err := c.connection.ReadMessage()
		if err != nil {
			println("Input reading failed, player dropped")
			c.alive = false
			break
		}

		// TODO More sanitation?
		if len(msg) > 1000 {
			println("Too long message, not processing")
			continue
		}
		message := string(msg)
		fmt.Printf("%s sent: %s\n", c.connection.RemoteAddr(), message)

		var input = string(msg)
		if input == "up" && c.wantedDirection != down {
			c.wantedDirection = up
		} else if input == "left" && c.wantedDirection != right {
			c.wantedDirection = left
		} else if input == "down" && c.wantedDirection != up {
			c.wantedDirection = down
		} else if input == "right" && c.wantedDirection != left {
			c.wantedDirection = right
		}
	}
}

func broadcastGameState(gameState GameStateMessage) {
	var message, _ = json.Marshal(gameState)
	for i := 0; i < len(clients); i++ {
		if !clients[i].alive {
			continue
		}
		var err = sendMessageToClient(clients[i].connection, GameStateUpdate, message)
		if err != nil {
			closeError := clients[i].connection.Close()
			clients[i].alive = false
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
		if !client.alive {
			continue
		}

		var err = sendMessageToClient(client.connection, messageType, jsonMessage)
		if err != nil {
			println("Failed to send message")
		}
	}
}
