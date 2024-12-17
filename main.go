package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
)

/*
Bugs

TODO
- Remove clients when connection is dropped. Max number on init?
-

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

type Direction int

const (
	up    Direction = 0
	left            = 1
	down            = 2
	right           = 3
)

type Client struct {
	direction       Direction
	wantedDirection Direction
	snake           []TailSegment
	connection      *websocket.Conn
	alive           bool
	tailLength      int
}

type Pickup struct {
	x          int
	y          int
	pickupType int
}

type TailSegment struct {
	x int
	y int
}

func toTailPosition(tailSegment []TailSegment, tailLength int) []TailMessage {
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

type GameState struct {
	Players []PlayerMessage
	Pickups []PickupMessage
}

type GameInitRequest struct {
	LevelWidth  int
	LevelHeight int
}

type GameSetupMessage struct {
	LevelWidth  int
	LevelHeight int
}

var levelWidth = 50
var levelHeight = 50

var clients = make([]*Client, 0)
var pickups = make([]Pickup, 5)

var gameRunning = false

func main() {
	http.HandleFunc("/game", game)
	http.HandleFunc("/host", host)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "client.html")
	})
	http.ListenAndServe(":8080", nil)
}

func host(responseWriter http.ResponseWriter, request *http.Request) {
	println("Hosting")
	conn, _ := upgrader.Upgrade(responseWriter, request, nil)
	_, msg, _ := conn.ReadMessage()

	gameInitRequest := GameInitRequest{}
	json.Unmarshal(msg, &gameInitRequest)

	println(gameInitRequest.LevelWidth)
}

func game(responseWriter http.ResponseWriter, request *http.Request) {
	conn, _ := upgrader.Upgrade(responseWriter, request, nil)
	client := Client{
		direction:       down,
		wantedDirection: down,
		connection:      conn,
		alive:           true,
		snake:           make([]TailSegment, 100),
		tailLength:      5,
	}
	client.snake[0].x = (10 + 10*len(clients)) % levelWidth
	client.snake[0].y = 10
	clients = append(clients, &client)

	go inputLoop(&client)
	if gameRunning || len(clients) < 2 {
		return
	}
	gameRunning = true
	println("Game started")
	initGame()
	gameLoop()

}

func initGame() {
	println("Init game")

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

func broadcastGameState(gameState GameState) {
	for i := 0; i < len(clients); i++ {
		if !clients[i].alive {
			continue
		}
		var message, _ = json.Marshal(gameState)
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
