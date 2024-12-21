package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go_project/game"
	"go_project/requests"
	"net/http"
	"time"
)

/*
Bugs

TODO
-

479 bytes per meddelande med json

Pickup ideas
- Faster speed
- Slower speed
- Invisible
- Bombs


Anti cheat
Send some hash to show code is not modified. What about replay attacks?
*/

var SendWithBinary = true

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	connection *websocket.Conn
	connected  bool
	player     game.Player
}

type PlayerListUpdateMessage struct {
	Entries []PlayerListEntry
}

type PlayerListEntry struct {
	Name  string
	Color string
	Score int
}

type GameSetupMessage struct {
	LevelWidth  int
	LevelHeight int
}

var clients = make([]*Client, 20)

func main() {
	game.InitGame()
	go gameLoop()

	http.HandleFunc("/join", joinGame)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "client.html")
	})

	http.ListenAndServe(":8080", nil)
}

// Do this with channels instead?
func gameLoop() {
	for {
		activeClients := make([]*game.Player, 0)
		for i := 0; i < len(clients); i++ {
			if clients[i] != nil && clients[i].connected {
				activeClients = append(activeClients, &clients[i].player)
			}
		}
		gameState := game.Tick(activeClients)

		broadcastGameState(gameState)
		if gameState.ScoreChanged {
			broadcastPlayerlist(&clients)
		}

		time.Sleep(80 * time.Millisecond)
	}
}

/*
Note: CheckOrigin in Upgrader allows all connections.
In a production environment, make sure to validate the origin to avoid Cross-Site WebSocket Hijacking.
*/
func joinGame(responseWriter http.ResponseWriter, request *http.Request) {

	conn, _ := upgrader.Upgrade(responseWriter, request, nil)
	//sendByteMessageToClient(conn, GameSetup, []byte("aaa"))
	//sendGameSetupMessage(conn, )
	_, msg, msgError := conn.ReadMessage()

	if msgError != nil {
		println("Error when reading join message")
		return
	}

	gameJoinRequest := requests.GameJoinRequest{}
	json.Unmarshal(msg, &gameJoinRequest)

	client, err := createClient(conn, gameJoinRequest.SnakeColor, gameJoinRequest.SnakeName)
	if err != nil {
		println(err.Error())
		sendMessageToClient(conn, TextMessage, []byte("Cannot connect. Too many players"))
		conn.Close()
		return
	}
	//clients = append(clients, client)

	gameSetup := GameSetupMessage{LevelWidth: game.LevelWidth, LevelHeight: game.LevelHeight}
	if SendWithBinary {
		sendGameSetupMessage(client.connection, gameSetup)
	} else {
		outgoingMessage, _ := json.Marshal(gameSetup)
		sendMessageToClient(client.connection, GameSetup, outgoingMessage)
	}

	broadcastPlayerlist(&clients)
	go inputLoop(client)
	println("Game started")
}

func lol() {
	buf := new(bytes.Buffer)
	var num uint16 = 1234
	err := binary.Write(buf, binary.LittleEndian, num)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	fmt.Printf("% x", buf.Bytes())
}

func broadcastPlayerlist(clients *[]*Client) {
	playerListUpdate := PlayerListUpdateMessage{make([]PlayerListEntry, 0)}

	for i := 0; i < len(*clients); i++ {
		client := (*clients)[i]
		if client == nil || !client.connected {
			continue
		}
		playerListUpdate.Entries = append(playerListUpdate.Entries, PlayerListEntry{
			Name:  client.player.Name,
			Color: client.player.SnakeColor,
			Score: client.player.TailLength,
		})
	}
	broadcastMessageToActiveClients(clients, PlayerListUpdate, playerListUpdate)
}

func createClient(connection *websocket.Conn, snakeColor string, name string) (*Client, error) {
	index, err := getFirstFreeSlotIndex(clients)
	if err != nil {
		return nil, err
	}
	player := game.CreatePlayer(name, snakeColor)

	client := Client{
		connection: connection,
		connected:  true,
		player:     player,
	}

	clients[index] = &client
	return &client, nil
}

func getFirstFreeSlotIndex(clients []*Client) (int, error) {
	for i := 0; i < len(clients); i++ {
		if clients[i] == nil || clients[i].connected == false {
			return i, nil
		}
	}
	return -1, &ServerFullError{}
}

func inputLoop(c *Client) {
	println("Starting input loop")

	for {
		if c.connected == false {
			break
		}
		_, msg, err := c.connection.ReadMessage()
		if err != nil {
			println("Input reading failed, player dropped")
			c.connected = false
			//c.connection.Close()
			break
		}

		// TODO More sanitation
		if len(msg) > 100 {
			println("Too long message, not processing")
			continue
		}
		//message := string(msg)
		//fmt.Printf("%s sent: %s\n", c.connection.RemoteAddr(), message)

		var input = string(msg)
		game.SetWantedDirection(&c.player, input)
	}
}

func broadcastGameState(gameState game.GameStateMessage) {
	var message, _ = json.Marshal(gameState)
	for i := 0; i < len(clients); i++ {
		if clients[i] == nil {
			continue
		}
		if !clients[i].connected {
			clients[i].connection.Close()
			continue
		}
		var err = sendMessageToClient(clients[i].connection, GameStateUpdate, message)
		if err != nil {
			closeError := clients[i].connection.Close()
			clients[i].connected = false
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
		if client == nil || !client.connected {
			continue
		}

		var err = sendMessageToClient(client.connection, messageType, jsonMessage)
		if err != nil {
			println("Failed to send message")
		}
	}
}
