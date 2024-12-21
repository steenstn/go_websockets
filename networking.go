package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type byte
	Msg  []byte
}

type MessageType byte

const (
	TextMessage      MessageType = 0
	GameStateUpdate              = 1
	GameSetup                    = 2
	PlayerListUpdate             = 3
)

func sendGameSetupMessage(connection *websocket.Conn, message GameSetupMessage) {
	binaryMessage := makeBinaryMessage(message)
	connection.WriteMessage(websocket.BinaryMessage, binaryMessage)
}

func makeBinaryMessage(message GameSetupMessage) []byte {
	byteArray := make([]byte, 5)
	byteArray[0] = GameSetup
	byteArray[1] = byte(message.LevelWidth >> 8)
	byteArray[2] = byte(message.LevelWidth)
	byteArray[3] = byte(message.LevelHeight >> 8)
	byteArray[4] = byte(message.LevelHeight)
	return byteArray
}

// Websocket binary 81
/*

300
0000 0001 0010 1100
*/
// gprc?
func sendMessageToClient(connection *websocket.Conn, messageType MessageType, message []byte) error {
	var resultingMessage, _ = json.Marshal(Message{
		Type: byte(messageType),
		Msg:  message,
	})
	err := connection.WriteMessage(1, resultingMessage)
	if err != nil {
		fmt.Printf("Error when sending message to user: %s\n", err)
		return err
	}
	return nil
}

func sendByteMessageToClient(connection *websocket.Conn, messageType MessageType, message []byte) error {
	var resultingMessage = make([]byte, 1)
	resultingMessage[0] = byte(messageType)
	resultingMessage = append(resultingMessage, message...)
	connection.WriteMessage(websocket.BinaryMessage, resultingMessage)
	return nil
}
