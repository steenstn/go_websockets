package main

import (
	"fmt"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type byte
	Msg  []byte
}

type MessageType byte

type GameMessage interface {
	toByteArray() []byte
}

const (
	TextMessage      MessageType = 0
	GameStateUpdate  MessageType = 1
	GameSetup        MessageType = 2
	PlayerListUpdate MessageType = 3
	HighScoreUpdate  MessageType = 4
)

func sendMessage(connection *websocket.Conn, message GameMessage) error {
	binaryMessage := message.toByteArray()
	err := connection.WriteMessage(websocket.BinaryMessage, binaryMessage)
	if err != nil {
		fmt.Printf("Failed to send message: %s", err)
	}
	return err
}
