package main

import "encoding/json"

type Message struct {
	Type byte
	Msg  []byte
}

type MessageType byte

const (
	TextMessage    MessageType = 0
	PositionUpdate             = 1
)

func broadcastToPlayer(client *client, messageType MessageType, message []byte) {
	var resultingMessage, _ = json.Marshal(Message{
		Type: byte(messageType),
		Msg:  message,
	})
	err := client.connection.WriteMessage(1, resultingMessage)
	if err != nil {
		println("Error when sending message to user")
	}
}
