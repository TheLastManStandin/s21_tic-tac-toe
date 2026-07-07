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
	const p1 = 1 // всегда человек
	const p2 = 2
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

func (s Service) GameEnded() bool {
	return false
}
