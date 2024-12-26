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
