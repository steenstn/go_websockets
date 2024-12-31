package main

import (
	"github.com/stretchr/testify/assert"
	"go_project/game"
	"testing"
)

func TestTextMessage(t *testing.T) {
	message := TextInfoMessage{Text: "yeah"}
	binaryMessage := message.toByteArray()
	assert.Equal(t, 7, len(binaryMessage))
	assert.Equal(t, messageVersion, binaryMessage[0])
	assert.Equal(t, TextMessage, MessageType(binaryMessage[1]))
	assert.Equal(t, "y", string(binaryMessage[2]))
	assert.Equal(t, "e", string(binaryMessage[3]))
	assert.Equal(t, "a", string(binaryMessage[4]))
	assert.Equal(t, "h", string(binaryMessage[5]))
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
	//	assert.Equal(t, PlayerListUpdate, MessageType(binaryMessage[1]))
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
