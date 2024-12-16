package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type byte
	Msg  []byte
}

type MessageType byte

const (
	TextMessage     MessageType = 0
	GameStateUpdate             = 1
)

// gprc?
func sendMessageToClient(connection *websocket.Conn, messageType MessageType, message []byte) error {
	var resultingMessage, _ = json.Marshal(Message{
		Type: byte(messageType),
		Msg:  message,
	})
	err := connection.WriteMessage(1, resultingMessage)
	if err != nil {
		println("Error when sending message to user")
		return err
	}
	return nil
}
