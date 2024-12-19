package main

import (
	"math/rand"
)

type Pickup struct {
	x          int
	y          int
	pickupType int
}

type TailSegment struct {
	x int
	y int
}

type Direction int

const (
	up    Direction = 0
	left            = 1
	down            = 2
	right           = 3
)

var levelWidth = 80
var levelHeight = 60

func gameTick(clients []*Client) GameStateMessage {

	pickupPositions := make([]PickupMessage, len(pickups))
	clientPositions := make([]PlayerMessage, 0)

	// Update snakes
	for i := 0; i < len(clients); i++ {
		if clients[i].alive == false {
			continue
		}
		input := clients[i].wantedDirection
		if input == up && clients[i].direction != down {
			clients[i].direction = up
		} else if input == left && clients[i].direction != right {
			clients[i].direction = left
		} else if input == down && clients[i].direction != up {
			clients[i].direction = down
		} else if input == right && clients[i].direction != left {
			clients[i].direction = right
		}
		moveSnake(&clients[i].snake, clients[i].tailLength, clients[i].direction)
		checkCollisionsWithSnakes(clients[i])
		checkCollisionsWithPickups(clients[i])

		wrapAround(clients[i], levelWidth, levelHeight, 0)

		clientPositions = append(clientPositions, PlayerMessage{X: clients[i].snake[0].x,
			Y:         clients[i].snake[0].y,
			Direction: clients[i].direction,
			Color:     clients[i].snakeColor,
			Tail:      toTailMessage(clients[i].snake, clients[i].tailLength)})
	}

	// Update pickups
	for i := 0; i < len(pickups); i++ {
		pickupPositions[i].X = pickups[i].x
		pickupPositions[i].Y = pickups[i].y
	}

	gameState := GameStateMessage{
		Players: clientPositions,
		Pickups: pickupPositions,
	}
	return gameState
}

func moveSnake(snakePointer *[]TailSegment, tailLength int, direction Direction) {
	var snake = *snakePointer
	// Move the tail segments, following the segment before it
	for i := tailLength; i > 0; i-- {
		snake[i].x = snake[i-1].x
		snake[i].y = snake[i-1].y
	}

	// Move the head
	switch direction {
	case up:
		snake[0].y--
	case left:
		snake[0].x--
	case down:
		snake[0].y++
	case right:
		snake[0].x++
	}
}

func checkCollisionsWithSnakes(client *Client) {
	headX := client.snake[0].x
	headY := client.snake[0].y

	for i := 0; i < len(clients); i++ {
		if clients[i].alive == false {
			continue
		}
		snakeToCheck := &clients[i].snake
		for j := 1; j < clients[i].tailLength; j++ {
			if headX == (*snakeToCheck)[j].x && headY == (*snakeToCheck)[j].y {
				client.alive = false
			}
		}
	}
}

func checkCollisionsWithPickups(client *Client) {
	for i := 0; i < len(pickups); i++ {
		if client.snake[0].x == pickups[i].x && client.snake[0].y == pickups[i].y {
			// Grow snake
			client.tailLength++
			client.snake[client.tailLength].x = client.snake[client.tailLength-1].x
			client.snake[client.tailLength].y = client.snake[client.tailLength-1].y

			// Reposition pickup
			pickups[i].x = rand.Intn(2 + levelWidth - 4)
			pickups[i].y = rand.Intn(2 + levelHeight - 4)
		}
	}
}

func wrapAround(position *Client, xMax int, yMax int, buffer int) {
	if position.snake[0].x >= xMax+buffer {
		position.snake[0].x = -buffer
	} else if position.snake[0].x < -buffer {
		position.snake[0].x = xMax + buffer
	}

	if position.snake[0].y >= yMax+buffer {
		position.snake[0].y = -buffer
	} else if position.snake[0].y < -buffer {
		position.snake[0].y = yMax + buffer
	}
}
