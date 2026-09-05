package minimax

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
