package game

import (
	"math/rand"
)

type Pickup struct {
	x          int
	y          int
	pickupType byte
}

type TailSegment struct {
	x int
	y int
}

type TopSnake struct {
	Name  string
	Score int
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
	HighScore    TopSnake
}

type PickupMessage struct {
	X int
	Y int
}

type PlayerMessage struct {
	Color string
	Tail  []TailMessage
}

type TailMessage struct {
	X int
	Y int
}

type Direction byte

const (
	up    Direction = 0
	left            = 1
	down            = 2
	right           = 3
)

var LevelWidth = 80
var LevelHeight = 60

var pickups = make([]Pickup, 5)
var pickupPositions = make([]PickupMessage, len(pickups))
var topSnake = TopSnake{}

func InitGame() {
	println("Initiating game")

	for i := 0; i < len(pickups); i++ {
		pickups[i].pickupType = 0
		pickups[i].x = rand.Intn(3 + LevelWidth - 5)
		pickups[i].y = rand.Intn(3 + LevelHeight - 5)
	}
}

func HandleInput(player *Player, input string) {
	if input == "U" && player.wantedDirection != down {
		player.wantedDirection = up
	} else if input == "L" && player.wantedDirection != right {
		player.wantedDirection = left
	} else if input == "D" && player.wantedDirection != up {
		player.wantedDirection = down
	} else if input == "R" && player.wantedDirection != left {
		player.wantedDirection = right
	} else if input == "S" {
		if !player.alive {
			spawnPlayer(player)
		}
	}
}

func Tick(players []*Player) GameStateMessage {
	clientPositions := make([]PlayerMessage, 0)
	scoreChanged := false

	// Update snakes
	for i := 0; i < len(players); i++ {
		if players[i].alive == false {
			continue
		}
		input := players[i].wantedDirection
		if input == up && players[i].direction != down {
			players[i].direction = up
		} else if input == left && players[i].direction != right {
			players[i].direction = left
		} else if input == down && players[i].direction != up {
			players[i].direction = down
		} else if input == right && players[i].direction != left {
			players[i].direction = right
		}
		moveSnake(&players[i].snake, players[i].TailLength, players[i].direction)

		if hasCollidedWithAnotherSnake(players[i], players) {
			players[i].alive = false
		}

		scoreChanged = checkCollisionsWithPickups(players[i])

		if players[i].TailLength > topSnake.Score {
			topSnake.Score = players[i].TailLength
			topSnake.Name = players[i].Name
		}

		if isOutOfBounds(players[i], LevelWidth, LevelHeight) {
			players[i].alive = false
		}

		clientPositions = append(clientPositions, PlayerMessage{
			Color: players[i].SnakeColor,
			Tail:  toTailMessage(players[i].snake, players[i].TailLength)})
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
		HighScore:    topSnake,
	}
	return gameState
}

func CreatePlayer(name string, color string) Player {
	player := Player{
		direction:       down,
		wantedDirection: down,
		snake:           make([]TailSegment, 255),
		alive:           true,
		TailLength:      3,
		SnakeColor:      color,
		Name:            name,
	}

	spawnPlayer(&player)
	return player
}

func spawnPlayer(player *Player) {
	player.direction = down
	player.wantedDirection = down
	player.alive = true
	player.TailLength = 3
	player.snake[0].x = 20 + rand.Intn(LevelWidth-40)
	player.snake[0].y = 10
	for i := 1; i < player.TailLength; i++ {
		player.snake[i].x = player.snake[0].x
		player.snake[i].y = player.snake[i-1].y - 1
	}
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

func hasCollidedWithAnotherSnake(player *Player, allPlayers []*Player) bool {
	headX := player.snake[0].x
	headY := player.snake[0].y

	for i := 0; i < len(allPlayers); i++ {
		if allPlayers[i] == nil || allPlayers[i].alive == false {
			continue
		}
		snakeToCheck := &allPlayers[i].snake
		for j := 1; j < allPlayers[i].TailLength; j++ {
			if headX == (*snakeToCheck)[j].x && headY == (*snakeToCheck)[j].y {
				return true
			}
		}
	}
	return false
}

func checkCollisionsWithPickups(player *Player) bool {
	for i := 0; i < len(pickups); i++ {
		if player.snake[0].x == pickups[i].x && player.snake[0].y == pickups[i].y {
			// Grow snake
			if player.TailLength < 255 {
				player.TailLength++
				player.snake[player.TailLength].x = player.snake[player.TailLength-1].x
				player.snake[player.TailLength].y = player.snake[player.TailLength-1].y
			}

			// Reposition pickup
			pickups[i].x = rand.Intn(2 + LevelWidth - 4)
			pickups[i].y = rand.Intn(2 + LevelHeight - 4)
			return true
		}
	}
	return false
}

func isOutOfBounds(position *Player, xMax int, yMax int) bool {
	return position.snake[0].x >= xMax || position.snake[0].x < 0 || position.snake[0].y >= yMax || position.snake[0].y < 0
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
