package handlers

type TurnRequest struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

type GameResponse struct {
	Table [3][3]int `json:"table"`
	State string    `json:"state"`
}

type RobotRequest struct {
	WithRobot bool `json:"WithRobot"`
}
