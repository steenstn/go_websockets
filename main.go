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
}

type clientMessage struct {
	X int
	Y int
}

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
	conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
	client := client{50 + 50*len(clients), 50, Down, conn}
	clients = append(clients, client)

	go inputLoop(len(clients) - 1)
	if gameRunning {
		return
	}
	gameRunning = true
	println("Game started")
	gameLoop()

}

func inputLoop(index int) {
	println("Starting input loop")
	for {
		_, msg, _ := clients[index].connection.ReadMessage()
		fmt.Printf("%s sent: %s\n", clients[index].connection.RemoteAddr(), string(msg))
		var input = string(msg)
		if input == "up" {
			clients[index].direction = Up
		} else if input == "left" {
			clients[index].direction = Left
		} else if input == "down" {
			clients[index].direction = Down
		} else if input == "right" {
			clients[index].direction = Right
		}
	}
}

func gameLoop() {
	for {
		clientPositions := make([]clientMessage, 0)
		for i := 0; i < len(clients); i++ {
			//	go readInput(&clients[i])
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
			clientPositions = append(clientPositions, clientMessage{clients[i].x, clients[i].y})

			// TODO: Send stuff as bytes instead of json strings
			//var result = "{\"x\":" + strconv.Itoa(clients[i].x) + ",\"y\":" + strconv.Itoa(clients[i].y) + "}"
			//	var result = clientMessage{clients[i].x, clients[i].y}
			//	var lol, _ = json.Marshal(result)
			//	clients[i].connection.WriteMessage(1, []byte(lol))
			//clients[i].connection.WriteMessage(1, []byte(result))
		}
		for i := 0; i < len(clients); i++ {
			//	var result = clientMessage{clients[i].x, clients[i].y}
			var lol, _ = json.Marshal(clientPositions)
			clients[i].connection.WriteMessage(1, lol)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
func handleMessage(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))
		conn.WriteMessage(msgType, msg)
		if string(msg) == "register" {
			println("aaw yeah")
			//	_ = append(clients, conn)
		}
		//println(clients)
		/*
			if string(msg) == "register" {
				println("aaw yeah")
				_ = append(clients, conn)
				//			return
			} else {
				println(clients)
				for i := 0; i < len(clients); i++ {
					clients[i].WriteMessage(msgType, msg)
				}
				//echo(w, r)
			}*/
	}
}
