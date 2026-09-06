package web

type GameRequest struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

type GameResponse struct {
	Table  [3][3]int `json:"table"`
	UUID   string    `json:"uuid"`
	Result string    `json:"result"`
}

type SignUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
