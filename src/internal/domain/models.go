package domain

type GameInstance struct {
	Table            [3][3]int
	State            GameState
	UUID             string
	Minimax          bool
	FirstPlayerUUID  string
	SecondPlayerUUID string
}

type GameState int32

const (
	WaitingForPlayers GameState = iota
	FirstPlayerTurn
	SecondPlayerTurn
	FirstPlayerWon
	SecondPlayerWon
	Draw
)

type UserInstance struct {
	Username string
	Password string
	UUID     string
}

type JwtRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type JwtResponse struct {
	Type         string `json:"type"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
