package domain

type GameService interface {
	MinimaxTurn() (GameInstance, error)
	Validate() error
	EndGame() int
}
type GameInstance struct {
	UUID   string
	Table  [3][3]int
	Status MinimaxStatus
}

type GameRequest struct {
	X int
	Y int
}

type MinimaxStatus int

const (
	MinimaxWin  MinimaxStatus = 2
	Draw        MinimaxStatus = 1
	MinimaxLose MinimaxStatus = 0
	Still       MinimaxStatus = -1
)

type minimaxGame struct {
	table      [3][3]int
	worstMarch MinimaxStatus
}
