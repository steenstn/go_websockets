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

type client struct {
	x          int
	y          int
	direction  int
	connection *websocket.Conn
	alive      bool
}

type clientMessage struct {
	X int
	Y int
}

const levelWidth = 320
const levelHeight = 240

var level = [levelWidth][levelHeight]int{}
var clients = make([]client, 0)
var gameRunning = false

func main() {
	http.HandleFunc("/echo", game)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websockets.html")
	})
	http.ListenAndServe(":8080", nil)
}

func game(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	client := client{50 + 50*len(clients), 50, Down, conn, true}
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
		if clients[index].alive == false {
			break
		}
		_, msg, err := clients[index].connection.ReadMessage()
		if err != nil {
			println("Input reading failed, player dropped")
			break
		}
		fmt.Printf("%s sent: %s\n", clients[index].connection.RemoteAddr(), string(msg))
		var input = string(msg)
		if input == "up" && clients[index].direction != Down {
			clients[index].direction = Up
		} else if input == "left" && clients[index].direction != Right {
			clients[index].direction = Left
		} else if input == "down" && clients[index].direction != Up {
			clients[index].direction = Down
		} else if input == "right" && clients[index].direction != Left {
			clients[index].direction = Right
		}
	}
}

func gameLoop() {
	for {
		clientPositions := make([]clientMessage, 0)
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
			} else {
				level[clients[i].x][clients[i].y] = 1
				clientPositions = append(clientPositions, clientMessage{clients[i].x, clients[i].y})
			}
			broadcastToPlayers(clientPositions)
			// TODO: Send stuff as bytes instead of json strings
			//var result = "{\"x\":" + strconv.Itoa(clients[i].x) + ",\"y\":" + strconv.Itoa(clients[i].y) + "}"
			//	var result = clientMessage{clients[i].x, clients[i].y}
			//	var lol, _ = json.Marshal(result)
			//	clients[i].connection.WriteMessage(1, []byte(lol))
			//clients[i].connection.WriteMessage(1, []byte(result))
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func isOutsideLevel(client *client) bool {
	if client.x >= levelWidth || client.x < 0 || client.y >= levelHeight || client.y < 0 {
		return true
	}
	return false
}

func broadcastToPlayers(messages []clientMessage) {
	for i := 0; i < len(clients); i++ {
		//	var result = clientMessage{clients[i].x, clients[i].y}
		var lol, _ = json.Marshal(messages)
		clients[i].connection.WriteMessage(1, lol)
	}
}
