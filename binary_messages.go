package main

import "go_project/game"

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
	byteArray[1] = byte(GameSetup)
	byteArray[2] = byte(message.LevelWidth)
	byteArray[3] = byte(message.LevelHeight)
	return byteArray
}

/*
0 - messageVersion
1 - message type
3..n text message
*/
func (message *TextInfoMessage) toByteArray() []byte {
	messageLength := len(message.Text)
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

func (message *GameStateMessageWrapper) toByteArray() []byte {

	/*
		gameState := message.state
		headerSize := 2
		totalPickupSize := 2 * len(gameState.Pickups) // X and Y for each pickup

		for playerIndex := 0; playerIndex < len(gameState.Players); playerIndex++ {

		}
	*/
	/*
		   message.state.ScoreChanged bool
		   message.state.Pickups[]
				X Y
		   message.state.Players[]
				Color
				Tail[]
					X
					Y
	*/
	return nil
}

/*
Get the corners plus endpoints from a continuous snake
*/
func getCorners(positions []game.TailMessage) []game.TailMessage {
	corners := make([]game.TailMessage, getCornerCount(&positions))
	if len(positions) == 0 {
		return corners
	}
	index := 0

	corners[index] = positions[0]
	index++

	for i := 1; i < len(positions)-1; i++ {
		a := positions[i-1]
		b := positions[i+1]
		if a.X != b.X && a.Y != b.Y {
			corners[index] = positions[i]
			index++
		}
	}

	corners[index] = positions[len(positions)-1]
	return corners
}

func getCornerCount(positions *[]game.TailMessage) int {
	numCorners := 0
	for i := 1; i < len(*positions)-1; i++ {
		a := (*positions)[i-1]
		b := (*positions)[i+1]
		if a.X != b.X && a.Y != b.Y {
			numCorners++
		}
	}
	numCorners += 2 // Add start and end point
	return numCorners
}

/*
0 - message version
1 - message type
2..n array of entities

In one entity:
1 byte - name length
x bytes - name
7 bytes - color (Hexadecimal #FFFFFF format)
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
	allScoreSize := 2 * numEntries             // 2 bytes for the score, score for every player
	allColorSizes := 7 * numEntries            // 7 bytes for the color (#abc123), color for every player

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

/*
0 - message version
1 - message type
2-3 - Score
4..n - Name
*/
func (message *HighScoreMessage) toByteArray() []byte {
	nameLength := len([]rune(message.Name))
	headerSize := 2
	scoreSize := 2
	byteArray := make([]byte, nameLength+headerSize+scoreSize)
	byteArray[0] = messageVersion
	byteArray[1] = byte(HighScoreUpdate)
	byteArray[2] = byte(message.Score >> 8)
	byteArray[3] = byte(message.Score)
	for i := 0; i < nameLength; i++ {
		byteArray[4+i] = message.Name[i]
	}
	return byteArray
}
