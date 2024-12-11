package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"math"
	"math/rand"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	x          float64
	y          float64
	vx         float64
	vy         float64
	direction  float64
	speed      float64
	connection *websocket.Conn
	alive      bool
}

type Asteroid struct {
	x         float64
	y         float64
	direction float64
	speed     float64
	size      int
}

type Bullet struct {
	x         float64
	y         float64
	direction float64
}

type PlayerPosition struct {
	X         int
	Y         int
	Direction float64
}

type AsteroidPosition struct {
	X    int
	Y    int
	Size int
}
type BulletPosition struct {
	X int
	Y int
}

type GameState struct {
	Players   []PlayerPosition
	Asteroids []AsteroidPosition
	Bullets   []BulletPosition
}

const levelWidth = 640
const levelHeight = 480

var clients = make([]Client, 0)
var asteroids = make([]Asteroid, 0)
var bullets = make([]Bullet, 0)

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
	client := Client{float64(50.0 + 50*len(clients)), 50.0, 0.0, 0.0, 0.0, 0.0, conn, true}
	clients = append(clients, client)

	go inputLoop(len(clients) - 1)
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
	for i := 0; i < 3; i++ {
		asteroids = append(asteroids, Asteroid{
			x:         float64(100 + 100*i),
			y:         200.0,
			direction: rand.Float64() * 2 * math.Pi,
			speed:     rand.Float64() * 2,
			size:      int(rand.Float64()*4 + 1),
		})
	}
}

func inputLoop(index int) {
	println("Starting input loop")
	for {
		c := &clients[index] // Why can't this be outside the loop?
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
		if input == "up" {
			c.speed = 0.2
		} else if input == "left" {
			c.direction -= 0.1
		} else if input == "down" {
			c.speed = 0
		} else if input == "right" {
			c.direction += 0.1
		} else if input == "space" {
			bullets = append(bullets, Bullet{
				x:         c.x,
				y:         c.y,
				direction: c.direction,
			})
		}
	}
}

func gameLoop() {
	for {
		clientPositions := make([]PlayerPosition, 0)
		for i := 0; i < len(clients); i++ {
			if clients[i].alive == false {
				continue
			}

			clients[i].vx += clients[i].speed * math.Cos(clients[i].direction)
			clients[i].vy += clients[i].speed * math.Sin(clients[i].direction)
			clients[i].x += clients[i].vx
			clients[i].y += clients[i].vy

			if clients[i].x > levelWidth+10 {
				clients[i].x = -10
			}
			if clients[i].x < -10 {
				clients[i].x = levelWidth + 10
			}
			if clients[i].y < -10 {
				clients[i].y = levelHeight + 10
			}
			if clients[i].y > levelHeight+10 {
				clients[i].y = -10
			}
			/*if isOutsideLevel(&clients[i]) || level[clients[i].x][clients[i].y] == 1 {
				clients[i].alive = false
				sendMessageToClient(clients[i].connection, TextMessage, []byte("you ded"))
			} else {
				//level[clients[i].x][clients[i].y] = 1
			}*/
			clientPositions = append(clientPositions, PlayerPosition{int(math.Round(clients[i].x)), int(math.Round(clients[i].y)), clients[i].direction})
		}

		asteroidPositions := make([]AsteroidPosition, 0)
		for i := 0; i < len(asteroids); i++ {
			asteroids[i].x += asteroids[i].speed * math.Cos(asteroids[i].direction)
			asteroids[i].y += asteroids[i].speed * math.Sin(asteroids[i].direction)

			if asteroids[i].x > levelWidth+50 {
				asteroids[i].x = -50
			}
			if asteroids[i].x < -50 {
				asteroids[i].x = levelWidth + 50
			}
			if asteroids[i].y < -50 {
				asteroids[i].y = levelHeight + 50
			}
			if asteroids[i].y > levelHeight+50 {
				asteroids[i].y = -50
			}

			asteroidPositions = append(asteroidPositions, AsteroidPosition{
				X:    int(math.Round(asteroids[i].x)),
				Y:    int(math.Round(asteroids[i].y)),
				Size: asteroids[i].size,
			})
		}

		bulletPositions := make([]BulletPosition, 0)
		for i := 0; i < len(bullets); i++ {
			bullets[i].x += 2 * math.Cos(bullets[i].direction)
			bullets[i].y += 2 * math.Sin(bullets[i].direction)
			bulletPositions = append(bulletPositions, BulletPosition{
				X: int(bullets[i].x),
				Y: int(bullets[i].y),
			})
		}

		gameState := GameState{
			Players:   clientPositions,
			Asteroids: asteroidPositions,
			Bullets:   bulletPositions,
		}

		broadcastGameState(gameState)
		time.Sleep(30 * time.Millisecond)
	}
}

func broadcastGameState(gameState GameState) {
	for i := 0; i < len(clients); i++ {
		var message, _ = json.Marshal(gameState)
		sendMessageToClient(clients[i].connection, GameStateUpdate, message)
	}
}
