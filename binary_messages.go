package main

var messageVersion byte = 1

func (message *GameSetupMessage) toByteArray() []byte {
	byteArray := make([]byte, 4)
	byteArray[0] = messageVersion
	byteArray[1] = GameSetup
	byteArray[2] = byte(message.LevelWidth)
	byteArray[3] = byte(message.LevelHeight)
	return byteArray
}

func (message *TextInfoMessage) toByteArray() []byte {
	var messageLength = len(message.Text)
	byteArray := make([]byte, 1+messageLength)
	byteArray[0] = messageVersion
	index := 1
	for i := 0; i < messageLength; i++ {
		byteArray[index] = message.Text[i]
		index++
	}
	return byteArray
}
