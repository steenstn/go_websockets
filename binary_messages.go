package main

const messageVersion byte = 1

/*
0 - messageVersion
1 - message type
2 - Level width
3 - Lebel height
*/
func (message *GameSetupMessage) toByteArray() []byte {
	byteArray := make([]byte, 4)
	byteArray[0] = messageVersion
	byteArray[1] = GameSetup
	byteArray[2] = byte(message.LevelWidth)
	byteArray[3] = byte(message.LevelHeight)
	return byteArray
}

/*
0 - version
1 - string length
2..n - message
*/
func (message *TextInfoMessage) toByteArray() []byte {
	var messageLength = len(message.Text)
	byteArray := make([]byte, 3+messageLength)
	byteArray[0] = messageVersion
	byteArray[1] = byte(TextMessage)
	index := 2
	for i := 0; i < messageLength; i++ {
		byteArray[index] = message.Text[i]
		index++
	}
	return byteArray
}
