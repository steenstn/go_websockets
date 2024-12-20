package game

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

type Player struct {
	direction       Direction
	wantedDirection Direction
	snake           []TailSegment
	alive           bool
	TailLength      int
	SnakeColor      string
	Name            string
}

type GameStateMessage struct {
	Players      []PlayerMessage
	Pickups      []PickupMessage
	ScoreChanged bool
}

type PickupMessage struct {
	X    int
	Y    int
	Type int
}

type PlayerMessage struct {
	X         int
	Y         int
	Direction Direction
	Color     string
	Tail      []TailMessage
}

type TailMessage struct {
	X int
	Y int
}

type Direction int

const (
	up    Direction = 0
	left            = 1
	down            = 2
	right           = 3
)

var LevelWidth = 80
var LevelHeight = 60

var pickups = make([]Pickup, 5)

func InitGame() {
	println("Initiating game")

	for i := 0; i < len(pickups); i++ {
		pickups[i].pickupType = 0
		pickups[i].x = rand.Intn(2 + LevelWidth - 4)
		pickups[i].y = rand.Intn(2 + LevelHeight - 4)
	}
}

func SetWantedDirection(player *Player, input string) {
	if input == "U" && player.wantedDirection != down {
		player.wantedDirection = up
	} else if input == "L" && player.wantedDirection != right {
		player.wantedDirection = left
	} else if input == "D" && player.wantedDirection != up {
		player.wantedDirection = down
	} else if input == "R" && player.wantedDirection != left {
		player.wantedDirection = right
	}
}

func Tick(clients []*Player) GameStateMessage {

	pickupPositions := make([]PickupMessage, len(pickups))
	clientPositions := make([]PlayerMessage, 0)
	scoreChanged := false
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
		moveSnake(&clients[i].snake, clients[i].TailLength, clients[i].direction)
		checkCollisionsWithSnakes(clients[i], clients)
		scoreChanged = checkCollisionsWithPickups(clients[i])

		wrapAround(clients[i], LevelWidth, LevelHeight, 0)

		clientPositions = append(clientPositions, PlayerMessage{X: clients[i].snake[0].x,
			Y:         clients[i].snake[0].y,
			Direction: clients[i].direction,
			Color:     clients[i].SnakeColor,
			Tail:      toTailMessage(clients[i].snake, clients[i].TailLength)})
	}

	// Update pickups
	for i := 0; i < len(pickups); i++ {
		pickupPositions[i].X = pickups[i].x
		pickupPositions[i].Y = pickups[i].y
	}

	gameState := GameStateMessage{
		Players:      clientPositions,
		Pickups:      pickupPositions,
		ScoreChanged: scoreChanged,
	}
	return gameState
}

func CreatePlayer(name string, color string) Player {
	player := Player{
		direction:       down,
		wantedDirection: down,
		snake:           make([]TailSegment, 100),
		alive:           true,
		TailLength:      5,
		SnakeColor:      color,
		Name:            name,
	}

	player.snake[0].x = 20 //(10 + 10*len(clients)) % levelWidth
	player.snake[0].y = 10
	return player
}

func toTailMessage(tailSegment []TailSegment, tailLength int) []TailMessage {
	result := make([]TailMessage, tailLength)
	for i := 0; i < tailLength; i++ {
		result[i].X = tailSegment[i].x
		result[i].Y = tailSegment[i].y
	}
	return result
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

func checkCollisionsWithSnakes(client *Player, clients []*Player) {
	headX := client.snake[0].x
	headY := client.snake[0].y

	for i := 0; i < len(clients); i++ {
		if clients[i] == nil || clients[i].alive == false {
			continue
		}
		snakeToCheck := &clients[i].snake
		for j := 1; j < clients[i].TailLength; j++ {
			if headX == (*snakeToCheck)[j].x && headY == (*snakeToCheck)[j].y {
				client.alive = false
			}
		}
	}
}

func checkCollisionsWithPickups(client *Player) bool {
	for i := 0; i < len(pickups); i++ {
		if client.snake[0].x == pickups[i].x && client.snake[0].y == pickups[i].y {
			// Grow snake
			client.TailLength++
			client.snake[client.TailLength].x = client.snake[client.TailLength-1].x
			client.snake[client.TailLength].y = client.snake[client.TailLength-1].y

			// Reposition pickup
			pickups[i].x = rand.Intn(2 + LevelWidth - 4)
			pickups[i].y = rand.Intn(2 + LevelHeight - 4)
			return true
		}
	}
	return false
}

func wrapAround(position *Player, xMax int, yMax int, buffer int) {
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
