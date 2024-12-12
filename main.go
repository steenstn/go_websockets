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
	position   Movable
	vx         float64
	vy         float64
	connection *websocket.Conn
	alive      bool
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
	client := Client{Movable{
		x:         float64(50.0 + 50*len(clients)),
		y:         50.0,
		direction: 0,
		speed:     0,
	}, 0.0, 0.0, conn, true}
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
	for i := 0; i < 5; i++ {
		asteroids = append(asteroids, Asteroid{position: Movable{
			x:         float64(100 + 100*i),
			y:         200.0,
			direction: rand.Float64() * 2 * math.Pi,
			speed:     rand.Float64() * 2},
			size: int(rand.Float64()*5 + 1),
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
			c.position.speed = 0.2
		} else if input == "left" {
			c.position.direction -= 0.15
		} else if input == "down" {
			c.position.speed = 0
		} else if input == "right" {
			c.position.direction += 0.15
		} else if input == "space" {
			bullets = append(bullets, Bullet{Movable{
				x:         c.position.x,
				y:         c.position.y,
				direction: c.position.direction,
				speed:     math.Sqrt(c.vx*c.vx+c.vy*c.vy) + 1.0,
			}, true})
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

			clients[i].vx += clients[i].position.speed * math.Cos(clients[i].position.direction)
			clients[i].vy += clients[i].position.speed * math.Sin(clients[i].position.direction)
			clients[i].position.x += clients[i].vx
			clients[i].position.y += clients[i].vy
			wrapAround(&clients[i].position, levelWidth, levelHeight, 10)

			clientPositions = append(clientPositions, PlayerPosition{int(math.Round(clients[i].position.x)), int(math.Round(clients[i].position.y)), clients[i].position.direction})
		}

		asteroidPositions := make([]AsteroidPosition, 0)
		for i := 0; i < len(asteroids); i++ {
			asteroids[i].position.x += asteroids[i].position.speed * math.Cos(asteroids[i].position.direction)
			asteroids[i].position.y += asteroids[i].position.speed * math.Sin(asteroids[i].position.direction)

			wrapAround(&asteroids[i].position, levelWidth, levelHeight, 50)
			for j := 0; j < len(clients); j++ {
				if overlap(asteroids[i].position.x, asteroids[i].position.y, float64(asteroids[i].size*9.0), float64(asteroids[i].size*9.0), clients[j].position.x, clients[j].position.y, 10, 10) {
					clients[j].alive = false
				}
			}

			asteroidPositions = append(asteroidPositions, AsteroidPosition{
				X:    int(math.Round(asteroids[i].position.x)),
				Y:    int(math.Round(asteroids[i].position.y)),
				Size: asteroids[i].size,
			})
		}

		bulletPositions := make([]BulletPosition, 0)
		for i := 0; i < len(bullets); i++ {
			if !bullets[i].alive {
				continue
			}
			currentBullet := &bullets[i].position
			currentBullet.x += currentBullet.speed * math.Cos(currentBullet.direction)
			currentBullet.y += currentBullet.speed * math.Sin(currentBullet.direction)

			for j := 0; j < len(asteroids); j++ {
				if asteroids[j].size > 0 && overlap(currentBullet.x, currentBullet.y, 6, 6, asteroids[j].position.x, asteroids[j].position.y, float64(asteroids[j].size*10.0), float64(asteroids[j].size*10.0)) {
					asteroids[j].size--
					bullets[i].alive = false
					if asteroids[j].size > 0 {
						asteroids[j].position.direction = rand.Float64() * 2 * math.Pi
						asteroids = append(asteroids, Asteroid{
							position: Movable{asteroids[j].position.x, asteroids[j].position.y, rand.Float64() * 2 * math.Pi, asteroids[j].position.speed * 1.2},
							size:     asteroids[j].size,
							alive:    true,
						})
					}
				}
			}

			bulletPositions = append(bulletPositions, BulletPosition{
				X: int(currentBullet.x),
				Y: int(currentBullet.y),
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

func wrapAround(position *Movable, xMax float64, yMax float64, buffer float64) {
	if position.x > xMax+buffer {
		position.x = -buffer
	}
	if position.x < -buffer {
		position.x = xMax + buffer
	}
	if position.y < -buffer {
		position.y = yMax + buffer
	}
	if position.y > yMax+buffer {
		position.y = -buffer
	}
}

func overlap(x float64, y float64, width float64, height float64, x2 float64, y2 float64, width2 float64, height2 float64) bool {
	return x < x2+width2 && x+width > x2 && y < y2+height2 && y+height > y2
}

func broadcastGameState(gameState GameState) {
	for i := 0; i < len(clients); i++ {
		var message, _ = json.Marshal(gameState)
		sendMessageToClient(clients[i].connection, GameStateUpdate, message)
	}
}
