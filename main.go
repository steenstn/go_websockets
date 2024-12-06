package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
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
	x          int
	y          int
	direction  int
	connection *websocket.Conn
	alive      bool
}

type PlayerPosition struct {
	X int
	Y int
}

const levelWidth = 320
const levelHeight = 240

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
	client := Client{50 + 50*len(clients), 50, Down, conn, true}
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
		if input == "up" && c.direction != Down {
			c.direction = Up
		} else if input == "left" && c.direction != Right {
			c.direction = Left
		} else if input == "down" && c.direction != Up {
			c.direction = Down
		} else if input == "right" && c.direction != Left {
			c.direction = Right
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
			switch clients[i].direction {
			case Up:
				println("up")
				clients[i].y--
			case Left:
				println("left")
				clients[i].x--
			case Down:
				println("down")
				clients[i].y++
			case Right:
				println("right")
				clients[i].x++
			}
			if isOutsideLevel(&clients[i]) || level[clients[i].x][clients[i].y] == 1 {
				clients[i].alive = false
				sendMessageToClient(clients[i].connection, TextMessage, []byte("you ded"))
			} else {
				level[clients[i].x][clients[i].y] = 1
				clientPositions = append(clientPositions, PlayerPosition{clients[i].x, clients[i].y})
			}
		}
		broadcastToPlayers(clientPositions)
		time.Sleep(200 * time.Millisecond)
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
