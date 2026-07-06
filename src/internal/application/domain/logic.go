package domain

func (s Service) HandleTurn(game Game) (Game, error) {
	s.ValidateGame(game)
	// Validate

	// minimax / player logic
	//
	return Game{}, nil
}

func (s Service) MinimaxNextTurn() {

}

func (s Service) ValidateGame(game Game) (error, bool) {
	return nil, false
}

func (s Service) GameEnded() bool {
	return false
}
