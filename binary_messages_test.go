package main

import (
	"github.com/stretchr/testify/assert"
	"go_project/game"
	"testing"
)

func TestTextMessage(t *testing.T) {
	message := TextInfoMessage{Text: "yeah"}
	binaryMessage := message.toByteArray()
	assert.Equal(t, 6, len(binaryMessage))
	assert.Equal(t, messageVersion, binaryMessage[0])
	assert.Equal(t, TextMessage, MessageType(binaryMessage[1]))
	assert.Equal(t, "y", string(binaryMessage[2]))
	assert.Equal(t, "e", string(binaryMessage[3]))
	assert.Equal(t, "a", string(binaryMessage[4]))
	assert.Equal(t, "h", string(binaryMessage[5]))
}

func TestHighScoreMessage(t *testing.T) {
	message := HighScoreMessage{
		Name:  "someone",
		Score: 666,
	}
	binaryMessage := message.toByteArray()
	assert.Equal(t, 11, len(binaryMessage))
	assert.Equal(t, messageVersion, binaryMessage[0])
	assert.Equal(t, HighScoreUpdate, MessageType(binaryMessage[1]))
	assert.Equal(t, 666, (int(binaryMessage[2])<<8)+int(binaryMessage[3]))
	assert.Equal(t, "s", string(binaryMessage[4]))
	assert.Equal(t, "o", string(binaryMessage[5]))
	assert.Equal(t, "m", string(binaryMessage[6]))
	assert.Equal(t, "e", string(binaryMessage[7]))
	assert.Equal(t, "o", string(binaryMessage[8]))
	assert.Equal(t, "n", string(binaryMessage[9]))
	assert.Equal(t, "e", string(binaryMessage[10]))
}

func TestGameStateMessage(t *testing.T) {
	message := GameStateMessageWrapper{game.GameStateMessage{
		Players: []game.PlayerMessage{
			{
				Color: "#abc123",
				Tail: []game.TailMessage{
					{X: 1, Y: 1},
					{X: 1, Y: 2},
					{X: 1, Y: 3},
					{X: 2, Y: 3},
					{X: 3, Y: 3},
					{X: 3, Y: 4},
				},
			},
			{
				Color: "#def456",
				Tail: []game.TailMessage{
					{X: 4, Y: 4},
					{X: 5, Y: 4},
				},
			},
		},
		Pickups: []game.PickupMessage{
			{X: 1, Y: 2},
			{X: 3, Y: 4},
		},
		ScoreChanged: true,
		HighScore:    game.TopSnake{},
	}}

	binaryMessage := message.toByteArray()
	assert.Equal(t, 37, len(binaryMessage))
	assert.Equal(t, messageVersion, binaryMessage[0])
	assert.Equal(t, GameStateUpdate, MessageType(binaryMessage[1]))

	// flags
	assert.Equal(t, 1, int(binaryMessage[2]))
	// num players
	assert.Equal(t, 2, int(binaryMessage[3]))

	// Color of first player
	assert.Equal(t, "#", string(binaryMessage[4]))
	assert.Equal(t, "a", string(binaryMessage[5]))
	assert.Equal(t, "b", string(binaryMessage[6]))
	assert.Equal(t, "c", string(binaryMessage[7]))
	assert.Equal(t, "1", string(binaryMessage[8]))
	assert.Equal(t, "2", string(binaryMessage[9]))
	assert.Equal(t, "3", string(binaryMessage[10]))

	// Number of corners and corner positions first player
	assert.Equal(t, 4, int(binaryMessage[11]))
	assert.Equal(t, 1, int(binaryMessage[12]))
	assert.Equal(t, 1, int(binaryMessage[13]))
	assert.Equal(t, 1, int(binaryMessage[14]))
	assert.Equal(t, 3, int(binaryMessage[15]))
	assert.Equal(t, 3, int(binaryMessage[16]))
	assert.Equal(t, 3, int(binaryMessage[17]))
	assert.Equal(t, 3, int(binaryMessage[18]))
	assert.Equal(t, 4, int(binaryMessage[19]))

	// Color of second player
	assert.Equal(t, "#", string(binaryMessage[20]))
	assert.Equal(t, "d", string(binaryMessage[21]))
	assert.Equal(t, "e", string(binaryMessage[22]))
	assert.Equal(t, "f", string(binaryMessage[23]))
	assert.Equal(t, "4", string(binaryMessage[24]))
	assert.Equal(t, "5", string(binaryMessage[25]))
	assert.Equal(t, "6", string(binaryMessage[26]))

	// Length and positions of second player
	assert.Equal(t, 2, int(binaryMessage[27]))
	assert.Equal(t, 4, int(binaryMessage[28]))
	assert.Equal(t, 4, int(binaryMessage[29]))
	assert.Equal(t, 5, int(binaryMessage[30]))
	assert.Equal(t, 4, int(binaryMessage[31]))

	// Pickups
	assert.Equal(t, 2, int(binaryMessage[32]))
	assert.Equal(t, 1, int(binaryMessage[33]))
	assert.Equal(t, 2, int(binaryMessage[34]))
	assert.Equal(t, 3, int(binaryMessage[35]))
	assert.Equal(t, 4, int(binaryMessage[36]))
}

func TestPlayerListUpdateMessage(t *testing.T) {
	message := PlayerListUpdateMessage{
		Entries: []PlayerListEntry{
			{
				Name:  "Test",
				Color: "#ABC123",
				Score: 500,
			},
		},
	}

	binaryMessage := message.toByteArray()
	// messageVersion + type + nameLengh + len(name) + color + score(2 bytes)
	assert.Equal(t, 16, len(binaryMessage))
	assert.Equal(t, messageVersion, binaryMessage[0])
	assert.Equal(t, PlayerListUpdate, MessageType(binaryMessage[1]))
	assert.Equal(t, 4, int(binaryMessage[2]))
	assert.Equal(t, "T", string(binaryMessage[3]))
	assert.Equal(t, "e", string(binaryMessage[4]))
	assert.Equal(t, "s", string(binaryMessage[5]))
	assert.Equal(t, "t", string(binaryMessage[6]))
	assert.Equal(t, "#", string(binaryMessage[7]))
	assert.Equal(t, "A", string(binaryMessage[8]))
	assert.Equal(t, "B", string(binaryMessage[9]))
	assert.Equal(t, "C", string(binaryMessage[10]))
	assert.Equal(t, "1", string(binaryMessage[11]))
	assert.Equal(t, "2", string(binaryMessage[12]))
	assert.Equal(t, "3", string(binaryMessage[13]))
	score := int(binaryMessage[14])<<8 + int(binaryMessage[15])
	assert.Equal(t, 500, score)
}

func TestGetCorners(t *testing.T) {
	snake := make([]game.TailMessage, 20)
	snake[0] = game.TailMessage{X: 1, Y: 1}
	snake[1] = game.TailMessage{X: 1, Y: 2}
	snake[2] = game.TailMessage{X: 1, Y: 3}
	snake[3] = game.TailMessage{X: 1, Y: 4}
	snake[4] = game.TailMessage{X: 2, Y: 4}
	snake[5] = game.TailMessage{X: 3, Y: 4}
	snake[6] = game.TailMessage{X: 4, Y: 4}
	snake[7] = game.TailMessage{X: 5, Y: 4}
	snake[8] = game.TailMessage{X: 5, Y: 5}
	snake[9] = game.TailMessage{X: 5, Y: 6}
	snake[10] = game.TailMessage{X: 6, Y: 6}
	snake[11] = game.TailMessage{X: 7, Y: 6}
	snake[12] = game.TailMessage{X: 8, Y: 6}
	snake[13] = game.TailMessage{X: 8, Y: 5}
	snake[14] = game.TailMessage{X: 8, Y: 4}
	snake[15] = game.TailMessage{X: 8, Y: 3}
	snake[16] = game.TailMessage{X: 7, Y: 3}
	snake[17] = game.TailMessage{X: 6, Y: 3}
	snake[18] = game.TailMessage{X: 6, Y: 2}
	snake[19] = game.TailMessage{X: 6, Y: 1}

	corners := getCorners(snake)
	assert.Equal(t, 8, len(corners))
	assert.Equal(t, game.TailMessage{1, 1}, corners[0])
	assert.Equal(t, game.TailMessage{1, 4}, corners[1])
	assert.Equal(t, game.TailMessage{5, 4}, corners[2])
	assert.Equal(t, game.TailMessage{5, 6}, corners[3])
	assert.Equal(t, game.TailMessage{8, 6}, corners[4])
	assert.Equal(t, game.TailMessage{8, 3}, corners[5])
	assert.Equal(t, game.TailMessage{6, 3}, corners[6])
	assert.Equal(t, game.TailMessage{6, 1}, corners[7])

}
