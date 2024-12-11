package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"math"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	Up    int = 0
	Left      = 1
	Down      = 2
	Right     = 3
)

type Client struct {
	x          float64
	y          float64
	direction  float64
	speed      float64
	connection *websocket.Conn
	alive      bool
}

type PlayerPosition struct {
	X int
	Y int
}

const levelWidth = 640
const levelHeight = 480

var level = [levelWidth][levelHeight]int{}
var clients = make([]Client, 0)
var gameRunning = false

func main() {
	http.HandleFunc("/game", game)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websockets.html")
	})
	http.ListenAndServe(":8080", nil)
}

func game(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	client := Client{float64(50.0 + 50*len(clients)), 50.0, 0.0, 0.0, conn, true}
	clients = append(clients, client)

	go inputLoop(len(clients) - 1)
	if gameRunning || len(clients) < 2 {
		return
	}
	gameRunning = true
	println("Game started")
	gameLoop()

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
			c.speed += 0.1
		} else if input == "left" {
			c.direction -= 0.1
		} else if input == "down" {
			//c.direction = Down
		} else if input == "right" {
			c.direction += 0.1
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
			clients[i].x += clients[i].speed * math.Cos(clients[i].direction)
			clients[i].y += clients[i].speed * math.Sin(clients[i].direction)

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
			clientPositions = append(clientPositions, PlayerPosition{int(math.Round(clients[i].x)), int(math.Round(clients[i].y))})
		}
		broadcastToPlayers(clientPositions)
		time.Sleep(30 * time.Millisecond)
	}
}

func isOutsideLevel(client *Client) bool {
	if client.x >= levelWidth || client.x < 0 || client.y >= levelHeight || client.y < 0 {
		return true
	}
	return false
}

func broadcastToPlayers(positions []PlayerPosition) {
	for i := 0; i < len(clients); i++ {
		var lol, _ = json.Marshal(positions)
		sendMessageToClient(clients[i].connection, PositionUpdate, lol)
	}
}
