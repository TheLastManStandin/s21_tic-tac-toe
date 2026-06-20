package api

type GameRequest struct {
	ID    string     `json:"id"`
	Board [3][3]int8 `json:"board"`
}
