package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"time"
)

/*
Bugs
The input is read "too often" so you if you are going up you can go left and then down before a server tick
which will cause you to turn back into yourself. Make a queue? add to queue if input is a new type of value
and read from it in gameloop

TODO
- Remove clients when connection is dropped. Max number on init?

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

func toTailPosition(tailSegment []TailSegment, tailLength int) []TailPosition {
	result := make([]TailPosition, tailLength)
	for i := 0; i < tailLength; i++ {
		result[i].X = tailSegment[i].x
		result[i].Y = tailSegment[i].y
	}
	return result
}

type PlayerPosition struct {
	X         int
	Y         int
	Direction Direction
	Tail      []TailPosition
}

type TailPosition struct {
	X int
	Y int
}

type PickupPosition struct {
	X    int
	Y    int
	Type int
}

type GameState struct {
	Players []PlayerPosition
	Pickups []PickupPosition
}

const levelWidth = 50
const levelHeight = 50

var clients = make([]*Client, 0)
var pickups = make([]Pickup, 5)

var gameRunning = false

func main() {
	http.HandleFunc("/game", game)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websockets.html")
	})
	http.ListenAndServe(":8080", nil)
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
			break
		}
		// TODO sanitize input
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

func gameLoop() {
	for {
		clientPositions := make([]PlayerPosition, 0)

		// Update snakes
		for i := 0; i < len(clients); i++ {
			if clients[i].alive == false {
				continue
			}
			input := clients[i].wantedDirection
			if input == up && clients[i].direction != down {
				clients[i].direction = up
			} else if input == left && clients[i].direction != right {
				clients[i].direction = left
			} else if input == down && clients[i].direction != up {
				clients[i].direction = down
			} else if input == right && clients[i].direction != left {
				clients[i].direction = right
			}
			moveSnake(&clients[i].snake, clients[i].tailLength, clients[i].direction)
			checkCollisionsWithSnakes(clients[i])
			checkCollisionsWithPickups(clients[i])

			wrapAround(clients[i], levelWidth, levelHeight, 1)

			clientPositions = append(clientPositions, PlayerPosition{clients[i].snake[0].x, clients[i].snake[0].y, clients[i].direction, toTailPosition(clients[i].snake, clients[i].tailLength)})
		}

		// Update pickups
		pickupPositions := make([]PickupPosition, len(pickups))
		for i := 0; i < len(pickups); i++ {
			pickupPositions[i].X = pickups[i].x
			pickupPositions[i].Y = pickups[i].y
		}

		gameState := GameState{
			Players: clientPositions,
			Pickups: pickupPositions,
		}

		broadcastGameState(gameState)
		time.Sleep(80 * time.Millisecond)
	}
}

func moveSnake(snakePointer *[]TailSegment, tailLength int, direction Direction) {
	var snake = *snakePointer
	// Move the tailsegments, following the segment before it
	for i := tailLength; i > 0; i-- {
		snake[i].x = snake[i-1].x
		snake[i].y = snake[i-1].y
	}

	// Move the head
	switch direction {
	case up:
		snake[0].y--
	case left:
		snake[0].x--
	case down:
		snake[0].y++
	case right:
		snake[0].x++
	}
}

func checkCollisionsWithSnakes(client *Client) {
	headX := client.snake[0].x
	headY := client.snake[0].y

	for i := 0; i < len(clients); i++ {
		if clients[i].alive == false {
			continue
		}
		snakeToCheck := &clients[i].snake
		for j := 1; j < clients[i].tailLength; j++ {
			if headX == (*snakeToCheck)[j].x && headY == (*snakeToCheck)[j].y {
				client.alive = false
				println("collision")
			}
		}
	}
}

func checkCollisionsWithPickups(client *Client) {
	for i := 0; i < len(pickups); i++ {
		if client.snake[0].x == pickups[i].x && client.snake[0].y == pickups[i].y {
			// Grow snake
			client.tailLength++
			client.snake[client.tailLength].x = client.snake[client.tailLength-1].x
			client.snake[client.tailLength].y = client.snake[client.tailLength-1].y

			// Reposition pickup
			pickups[i].x = rand.Intn(2 + levelWidth - 4)
			pickups[i].y = rand.Intn(2 + levelHeight - 4)
		}
	}
}

func wrapAround(position *Client, xMax int, yMax int, buffer int) {
	if position.snake[0].x > xMax+buffer {
		position.snake[0].x = -buffer
	}
	if position.snake[0].x < -buffer {
		position.snake[0].x = xMax + buffer
	}
	if position.snake[0].y < -buffer {
		position.snake[0].y = yMax + buffer
	}
	if position.snake[0].y > yMax+buffer {
		position.snake[0].y = -buffer
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
