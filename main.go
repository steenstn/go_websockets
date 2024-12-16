package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"time"
)

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
	x          int
	y          int
	direction  Direction
	snake      []TailSegment
	connection *websocket.Conn
	alive      bool
	tailLength int
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

type Asteroid struct {
	position Movable
	size     int
	alive    bool
}

type Movable struct {
	x         float64
	y         float64
	direction float64
	speed     float64
}

type Bullet struct {
	position Movable
	alive    bool
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
var pickups = make([]Pickup, 3)

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
		x:          10 + 15*len(clients),
		y:          5,
		direction:  down,
		connection: conn,
		alive:      true,
		snake:      make([]TailSegment, 100),
		tailLength: 5,
	}
	client.snake[0].x = 10 + 10*len(clients)
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
		fmt.Printf("%s sent: %s\n", c.connection.RemoteAddr(), string(msg))
		var input = string(msg)
		if input == "up" && c.direction != down {
			c.direction = up
		} else if input == "left" && c.direction != right {
			c.direction = left
		} else if input == "down" && c.direction != up {
			c.direction = down
		} else if input == "right" && c.direction != left {
			c.direction = right
		} else if input == "space" {
			/*bullets = append(bullets, Bullet{Movable{
				x:         c.position.x,
				y:         c.position.y,
				direction: c.position.direction,
				speed:     math.Sqrt(c.vx*c.vx+c.vy*c.vy) + 1.0,
			}, true})*/
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

			moveSnake(&clients[i].snake, clients[i].tailLength, clients[i].direction)
			checkCollisionsWithSnakes(clients[i])
			checkCollisionsWithPickups(clients[i])
			/*	clients[i].vy += clients[i].position.speed * math.Sin(clients[i].position.direction)
				clients[i].position.x += clients[i].vx
				clients[i].position.y += clients[i].vy*/
			wrapAround(clients[i], levelWidth, levelHeight, 1)

			clientPositions = append(clientPositions, PlayerPosition{clients[i].snake[0].x, clients[i].snake[0].y, clients[i].direction, toTailPosition(clients[i].snake, clients[i].tailLength)})
		}

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
	// Move the tail
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

func overlap(x float64, y float64, width float64, height float64, x2 float64, y2 float64, width2 float64, height2 float64) bool {
	return x < x2+width2 && x+width > x2 && y < y2+height2 && y+height > y2
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
