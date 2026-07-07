package api

type GameRequest struct {
	Table [3][3]int8 `json:"table"`
}
