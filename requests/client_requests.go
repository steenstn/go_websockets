package requests

import "regexp"

type GameJoinRequest struct {
	SnakeColor string
	SnakeName  string
}

func (q *GameJoinRequest) Validate() {
	r, _ := regexp.Compile("#[a-fA-F0-9]{6}")
	isValidColor := r.MatchString(q.SnakeColor)
	if !isValidColor {
		println("Invalid color")
		q.SnakeColor = "#FFFFFF"
	}
	if len(q.SnakeName) > 20 {
		q.SnakeName = q.SnakeName[:20]
	}
}
