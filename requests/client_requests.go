package requests

type GameInitRequest struct {
	LevelWidth  int
	LevelHeight int
	SnakeColor  string
	SnakeName   string
}

type GameJoinRequest struct {
	SnakeColor string
	SnakeJoin  string
	SnakeName  string
}
