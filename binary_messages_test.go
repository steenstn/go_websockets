package main

import (
	"github.com/stretchr/testify/assert"
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
