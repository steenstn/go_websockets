package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTextMessage(t *testing.T) {
	message := TextInfoMessage{Text: "yeah"}
	binaryMessage := message.toByteArray()
	assert.Equal(t, 5, len(binaryMessage))
	assert.Equal(t, messageVersion, binaryMessage[0])
	assert.Equal(t, byte('y'), binaryMessage[1])
	assert.Equal(t, byte('e'), binaryMessage[2])
	assert.Equal(t, byte('a'), binaryMessage[3])
	assert.Equal(t, byte('h'), binaryMessage[4])
}
