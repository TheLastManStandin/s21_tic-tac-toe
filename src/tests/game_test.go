package tests_test

import (
	"net/http"
	"slices"
	"testing"
)

func TestGameCreation(t *testing.T) {
	for _, withRobot := range []bool{false, true} {
		name, wantState := "with_player", "waiting_for_players"
		if withRobot {
			name, wantState = "with_robot", "first_player_turn"
		}
		t.Run(name, func(t *testing.T) {
			client := newAuthenticatedClient(t)
			id := client.createGame(t, withRobot)
			game := client.getGame(t, id)
			if game.UUID != id || game.FirstPlayerUUID != client.userID || game.Minimax != withRobot {
				t.Errorf("неверные параметры созданной игры: %+v", game)
			}
			if game.Table != [3][3]int{} || game.State != wantState {
				t.Errorf("ожидалось пустое поле и состояние %s, получено %+v", wantState, game)
			}
		})
	}
}

func TestGameRobotTurn(t *testing.T) {
	client := newAuthenticatedClient(t)
	id := client.createGame(t, true)
	game := client.makeTurn(t, id, 0, 0)
	if game.Table[0][0] != 1 {
		t.Fatal("ход пользователя не сохранён")
	}
	counts := [3]int{}
	for _, row := range game.Table {
		for _, cell := range row {
			if cell < 0 || cell > 2 {
				t.Fatalf("некорректное значение клетки: %d", cell)
			}
			counts[cell]++
		}
	}
	if counts[1] != 1 || counts[2] != 1 || game.State != "first_player_turn" {
		t.Errorf("после хода пользователя ожидался один ответный ход компьютера: %+v", game)
	}
	if saved := client.getGame(t, id); saved.Table != game.Table || saved.State != game.State {
		t.Error("результат хода не сохранён в игре")
	}
}

func TestGameJoining(t *testing.T) {
	owner := newAuthenticatedClient(t)
	guest := newAuthenticatedClient(t)
	id := owner.createGame(t, false)
	if !containsGame(guest.availableGames(t), id) {
		t.Fatal("созданная игра недоступна для подключения")
	}
	guest.joinGame(t, id)
	game := guest.getGame(t, id)
	if game.FirstPlayerUUID != owner.userID || game.SecondPlayerUUID != guest.userID || game.State != "first_player_turn" {
		t.Errorf("неверное состояние после подключения: %+v", game)
	}
	if containsGame(guest.availableGames(t), id) {
		t.Error("заполненная игра осталась в списке доступных")
	}
	guest.request(t, http.MethodPost, "/game/"+id+"/join", nil, http.StatusConflict)
}

func TestGameTurnValidation(t *testing.T) {
	owner := newAuthenticatedClient(t)
	guest := newAuthenticatedClient(t)
	outsider := newAuthenticatedClient(t)
	id := owner.createGame(t, false)
	t.Run("waiting_for_second_player", func(t *testing.T) {
		owner.request(t, http.MethodPost, "/game/"+id, turnRequest{X: 0, Y: 0}, http.StatusConflict)
	})
	guest.joinGame(t, id)
	before := owner.getGame(t, id)
	t.Run("wrong_player_turn", func(t *testing.T) {
		guest.request(t, http.MethodPost, "/game/"+id, turnRequest{X: 0, Y: 0}, http.StatusConflict)
	})
	t.Run("outside_board", func(t *testing.T) {
		owner.request(t, http.MethodPost, "/game/"+id, turnRequest{X: -1, Y: 0}, http.StatusBadRequest)
	})
	t.Run("outsider", func(t *testing.T) {
		outsider.request(t, http.MethodGet, "/game/"+id, nil, http.StatusForbidden)
		outsider.request(t, http.MethodPost, "/game/"+id, turnRequest{X: 0, Y: 0}, http.StatusForbidden)
	})
	if after := owner.getGame(t, id); after != before {
		t.Fatal("отклонённые запросы изменили игру")
	}
	owner.makeTurn(t, id, 0, 0)
	before = owner.getGame(t, id)
	t.Run("occupied_cell", func(t *testing.T) {
		guest.request(t, http.MethodPost, "/game/"+id, turnRequest{X: 0, Y: 0}, http.StatusConflict)
	})
	if after := owner.getGame(t, id); after != before {
		t.Error("попытка хода в занятую клетку изменила игру")
	}
}

func TestGameFinishedStates(t *testing.T) {
	for _, tc := range []struct {
		name  string
		moves []turnRequest
		state string
	}{
		{"first_player_wins", []turnRequest{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {0, 2}}, "first_player_won"},
		{"second_player_wins", []turnRequest{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {2, 2}, {1, 2}}, "second_player_won"},
		{"draw", []turnRequest{{0, 0}, {0, 1}, {0, 2}, {1, 1}, {1, 0}, {1, 2}, {2, 1}, {2, 0}, {2, 2}}, "draw"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			players := [2]*apiClient{newAuthenticatedClient(t), newAuthenticatedClient(t)}
			id := players[0].createGame(t, false)
			players[1].joinGame(t, id)
			var expected [3][3]int
			for i, move := range tc.moves {
				game := players[i%2].makeTurn(t, id, move.X, move.Y)
				expected[move.X][move.Y] = i%2 + 1
				wantState := "second_player_turn"
				if i%2 == 1 {
					wantState = "first_player_turn"
				}
				if i == len(tc.moves)-1 {
					wantState = tc.state
				}
				if game.Table != expected || game.State != wantState {
					t.Fatalf("ход %d: поле=%v состояние=%s; ожидалось поле=%v состояние=%s", i+1, game.Table, game.State, expected, wantState)
				}
			}
			before := players[0].getGame(t, id)
			if before.Table != expected || before.State != tc.state {
				t.Fatal("итог игры не сохранён")
			}
			// API может отклонить ход или вернуть уже завершённую игру.
			// В обоих случаях поле и результат должны остаться прежними.
			players[0].request(t, http.MethodPost, "/game/"+id, turnRequest{X: 2, Y: 0}, http.StatusOK, http.StatusConflict)
			if after := players[0].getGame(t, id); after != before {
				t.Error("ход после завершения изменил игру")
			}
		})
	}
}

func TestGameInvalidRequests(t *testing.T) {
	client := newAuthenticatedClient(t)
	t.Run("missing_game", func(t *testing.T) {
		client.request(t, http.MethodGet, "/game/"+missingUUID, nil, http.StatusNotFound)
	})
	t.Run("invalid_uuid", func(t *testing.T) {
		client.request(t, http.MethodGet, "/game/not-a-uuid", nil, http.StatusBadRequest)
	})
	t.Run("invalid_creation_body", func(t *testing.T) {
		client.request(t, http.MethodPost, "/game", `{"WithRobot":"yes"}`, http.StatusBadRequest)
	})
}

func TestGameHistory(t *testing.T) {
	t.Run("requires_access_token", func(t *testing.T) {
		client := newAPIClient(t)
		client.request(t, http.MethodGet, "/game/history", nil, http.StatusUnauthorized)
	})

	t.Run("empty_history", func(t *testing.T) {
		client := newAuthenticatedClient(t)
		if games := client.gameHistory(t); len(games) != 0 {
			t.Fatalf("у нового игрока должна быть пустая история, получено %+v", games)
		}
	})

	t.Run("finished_games_for_both_players_newest_first", func(t *testing.T) {
		first := newAuthenticatedClient(t)
		second := newAuthenticatedClient(t)
		third := newAuthenticatedClient(t)
		fourth := newAuthenticatedClient(t)

		finishGame := func(owner, guest *apiClient, moves []turnRequest, state string) gameResponse {
			t.Helper()
			id := owner.createGame(t, false)
			guest.joinGame(t, id)
			players := [2]*apiClient{owner, guest}
			var table [3][3]int
			for i, move := range moves {
				players[i%2].makeTurn(t, id, move.X, move.Y)
				table[move.X][move.Y] = i%2 + 1
			}
			return gameResponse{
				UUID: id, Table: table, State: state,
				FirstPlayerUUID: owner.userID, SecondPlayerUUID: guest.userID,
			}
		}

		firstWinMoves := []turnRequest{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {0, 2}}
		secondWinMoves := []turnRequest{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {2, 2}, {1, 2}}
		drawMoves := []turnRequest{{0, 0}, {0, 1}, {0, 2}, {1, 1}, {1, 0}, {1, 2}, {2, 1}, {2, 0}, {2, 2}}
		firstWon := finishGame(first, second, firstWinMoves, "first_player_won")
		foreign := finishGame(third, fourth, firstWinMoves, "first_player_won")
		secondWon := finishGame(first, second, secondWinMoves, "second_player_won")
		drawn := finishGame(first, second, drawMoves, "draw")

		first.createGame(t, false) // Игра в ожидании соперника.
		activeID := first.createGame(t, false)
		second.joinGame(t, activeID)
		first.makeTurn(t, activeID, 0, 0)

		for _, tc := range []struct {
			name   string
			client *apiClient
			want   []gameResponse
		}{
			{"first_player", first, []gameResponse{drawn, secondWon, firstWon}},
			{"second_player", second, []gameResponse{drawn, secondWon, firstWon}},
			{"unrelated_player", third, []gameResponse{foreign}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				if got := tc.client.gameHistory(t); !slices.Equal(got, tc.want) {
					t.Errorf("история = %+v, ожидалось %+v", got, tc.want)
				}
			})
		}
	})
}
