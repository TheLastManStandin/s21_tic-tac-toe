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

type GameDetailsResponse struct {
	UUID             string    `json:"uuid"`
	Table            [3][3]int `json:"table"`
	State            string    `json:"state"`
	Minimax          bool      `json:"minimax"`
	FirstPlayerUUID  string    `json:"firstPlayerUUID"`
	SecondPlayerUUID string    `json:"secondPlayerUUID"`
}
