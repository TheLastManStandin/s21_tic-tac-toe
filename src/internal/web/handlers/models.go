package handlers

type GameRequest struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

type GameResponse struct {
	Table [3][3]int `json:"table"`
	State string    `json:"state"`
}
