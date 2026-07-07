package domain

import (
	"errors"
	"strconv"
)

func (s Service) HandleTurn(game Game) (Game, error) {
	savedGame, err := s.db.GetGame(game.UUID)
	if err != nil {
		return Game{}, err
	}

	err = s.ValidateGameTurn(dbGameToGame(savedGame), game)
	if err != nil {
		return Game{}, err
	}

	// minimax / player logic
	// todo player 2 turn
	s.MinimaxNextTurn()

	return Game{}, nil
}

func (s Service) MinimaxNextTurn() {

}

func (s Service) ValidateGameTurn(savedGame Game, recivedGame Game) error {
	count := []int{0, 0, 0}
	hasChanged := 0

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			recivedVal := recivedGame.Table[i][j]
			savedVal := savedGame.Table[i][j]
			if recivedVal < 0 || recivedVal > 2 {
				return errors.New("unknown value:" + strconv.Itoa(int(recivedVal)))
			}
			count[recivedVal]++

			if recivedVal != savedVal {
				hasChanged++
			}
		}
	}
	if count[1]-count[2] > 1 || count[1]-count[2] < 0 {
		return errors.New("wrong player had turn")
	} else if hasChanged == 0 {
		return errors.New("no turn made")
	} else if hasChanged > 1 {
		return errors.New("too many turns")
	}

	return nil
}

// win: 1\2 ; have no space: -1 ; have space, noone win: 0
func (s Service) GameEnded(game Game) int {
	for i := 0; i < 3; i++ {
		rowVal := game.Table[i][0]
		rowSame := true
		for j := 0; j < 3; j++ {
			if game.Table[i][j] != rowVal {
				rowSame = false
			}
		}

		if rowSame && rowVal != 0 {
			return int(rowVal)
		}
	}

	for i := 0; i < 3; i++ {
		columnVal := game.Table[0][i]
		columnSame := true
		for j := 0; j < 3; j++ {
			if game.Table[j][i] != columnVal {
				columnSame = false
			}
		}

		if columnSame && columnVal != 0 {
			return int(columnVal)
		}
	}

	midVal := game.Table[1][1]
	if game.Table[0][0] == midVal && game.Table[2][2] == midVal && midVal != 0 {
		return int(midVal)
	} else if game.Table[0][2] == midVal && game.Table[2][0] == midVal && midVal != 0 {
		return int(midVal)
	}

	haveSpace := false
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if game.Table[i][j] == 0 {
				haveSpace = true
				break
			}
		}
	}
	if haveSpace {
		return 0
	} else {
		return -1
	}
}
