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

func (message *TextInfoMessage) toByteArray() []byte {
	var messageLength = len(message.Text)
	headerSize := 2
	byteArray := make([]byte, headerSize+messageLength)
	byteArray[0] = messageVersion
	byteArray[1] = byte(TextMessage)
	index := 2
	for i := 0; i < messageLength; i++ {
		byteArray[index] = message.Text[i]
		index++
	}
	return byteArray
}

/*
	type PlayerListUpdateMessage struct {
		Entries []PlayerListEntry
	}

	type PlayerListEntry struct {
		Name  string
		Color string
		Score int
	}
0 - message version
1 - message type
2..n array of entities

In one entity:
0 - name length
1..n - name

n+1..(n+1)+7 - color (#FFFFFF format)
(n+9)..n+11 - score (2 bytes)

1 byte - name length
x bytes - name
7 bytes - color (Hexadecimal #FFFFFF format )
2 bytes - score
*/

func (message *PlayerListUpdateMessage) toByteArray() []byte {
	numEntries := len(message.Entries)
	nameLengthSum := 0
	for i := 0; i < numEntries; i++ {
		nameLengthSum += len([]rune(message.Entries[i].Name))
	}
	headerSize := 2
	allNameSizes := numEntries + nameLengthSum // for each player: 1 byte for the string length and the string length itself
	allScoreSize := 2 * numEntries             // 2 bytes for the score
	allColorSizes := 7 * numEntries            // 7 bytes for the color (#abc123)
	byteArray := make([]byte, headerSize+allNameSizes+allScoreSize+allColorSizes)

	byteArray[0] = messageVersion
	byteArray[1] = byte(PlayerListUpdate)

	index := 2
	for i := 0; i < numEntries; i++ {
		byteArray[index] = byte(len([]rune(message.Entries[i].Name)))
		index++
		// Put name into byte array
		for nameIndex := 0; nameIndex < len([]rune(message.Entries[i].Name)); nameIndex++ {
			byteArray[index] = message.Entries[i].Name[nameIndex]
			index++
		}
		// Put color in byte array
		for colorIndex := 0; colorIndex < len([]rune(message.Entries[i].Color)); colorIndex++ {
			byteArray[index] = message.Entries[i].Color[colorIndex]
			index++
		}
		// Use two bytes to store the score
		byteArray[index] = byte(message.Entries[i].Score >> 8)
		index++
		byteArray[index] = byte(message.Entries[i].Score)
		index++
	}

	return byteArray
}
