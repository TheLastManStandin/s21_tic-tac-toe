package tests_test

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGameAgainstMinimax(t *testing.T) {
	tests := []struct {
		name      string
		moves     [][2]int
		wantState string
	}{
		{
			name:      "minimax wins",
			moves:     [][2]int{{0, 0}, {0, 1}, {1, 0}},
			wantState: "second_player_won",
		},
		{
			name:      "draw",
			moves:     [][2]int{{0, 0}, {0, 1}, {2, 0}, {1, 2}, {2, 2}},
			wantState: "draw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newAPIClient(t)

			// Prepare the game using the public API.
			api.request(t, http.MethodPost, "/register", nil)
			gameUUID := strings.TrimSpace(string(api.request(
				t, http.MethodPost, "/game", map[string]bool{"WithRobot": true},
			)))
			if gameUUID == "" {
				t.Fatal("create game returned an empty UUID")
			}

			var game gameResponse
			for turn, move := range tt.moves {
				x, y := move[0], move[1]
				if game.Table[x][y] != 0 {
					t.Fatalf("turn %d targets occupied cell (%d, %d): table = %v", turn+1, x, y, game.Table)
				}

				previousTable := game.Table
				game = api.game(t, http.MethodPost, gameUUID, turnRequest{X: x, Y: y})
				if game.Table[x][y] != 1 {
					t.Fatalf("player turn was not applied at (%d, %d): table = %v", x, y, game.Table)
				}
				assertValidTurn(t, previousTable, game.Table)

				if turn < len(tt.moves)-1 && gameFinished(game.State) {
					t.Fatalf("game finished before all planned turns: state = %q, table = %v", game.State, game.Table)
				}
			}

			if game.State != tt.wantState {
				t.Fatalf("final state = %q, want %q; table = %v", game.State, tt.wantState, game.Table)
			}

			// A separate HTTP request must return the saved game.
			savedGame := api.game(t, http.MethodGet, gameUUID, nil)
			if savedGame != game {
				t.Fatalf("saved game = %+v, last turn response = %+v", savedGame, game)
			}

			// Further moves must not change a finished game.
			finishedGame := api.game(t, http.MethodPost, gameUUID, turnRequest{X: 0, Y: 0})
			if finishedGame != game {
				t.Fatalf("finished game changed: before = %+v, after = %+v", game, finishedGame)
			}
			if savedGame := api.game(t, http.MethodGet, gameUUID, nil); savedGame != game {
				t.Fatalf("saved finished game changed: before = %+v, after = %+v", game, savedGame)
			}
		})
	}
}

func TestGameCreation(t *testing.T) {
	api := newAPIClient(t)
	userUUID := api.register(t)
	for _, withRobot := range []bool{false, true} {
		name := "multiplayer"
		if withRobot {
			name = "minimax"
		}
		t.Run(name, func(t *testing.T) {
			id := api.createGame(t, withRobot)
			game := api.gameDetails(t, id)
			if game.UUID != id || game.FirstPlayerUUID != userUUID || game.SecondPlayerUUID != "" || game.Minimax != withRobot || game.Table != ([3][3]int{}) {
				t.Fatalf("unexpected new game: %+v", game)
			}
			if !withRobot && game.State != "waiting_for_players" {
				t.Fatalf("new multiplayer game state = %q", game.State)
			}
		})
	}
	for _, tt := range []struct {
		name string
		body any
	}{
		{"missing field", `{}`},
		{"null field", `{"WithRobot":null}`},
		{"string instead of bool", `{"WithRobot":"true"}`},
		{"unknown field", `{"WithRobot":true,"other":false}`},
		{"malformed JSON", `{"WithRobot":`},
		{"multiple objects", `{"WithRobot":true}{"WithRobot":false}`},
		{"empty body", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			api.request(t, http.MethodPost, "/game", tt.body, http.StatusBadRequest)
		})
	}
}

func TestAvailableGamesAndJoining(t *testing.T) {
	owner, guest, outsider := newAPIClient(t), newAPIClient(t), newAPIClient(t)
	ownerUUID, guestUUID := owner.register(t), guest.register(t)
	outsider.register(t)
	waiting := owner.createGame(t, false)
	robot := owner.createGame(t, true)
	guestOwn := guest.createGame(t, false)

	t.Run("list shows joinable games", func(t *testing.T) {
		games := guest.availableGames(t)
		if !containsGame(games, waiting) {
			t.Fatal("available game missing from list")
		}
		if containsGame(games, robot) || containsGame(games, guestOwn) {
			t.Fatal("list contains robot game or own game")
		}
		for _, game := range games {
			if game.Minimax || game.State != "waiting_for_players" || game.SecondPlayerUUID != "" || game.FirstPlayerUUID == "" || game.FirstPlayerUUID == guestUUID {
				t.Fatalf("game is not joinable: %+v", game)
			}
		}
		if containsGame(owner.availableGames(t), waiting) {
			t.Fatal("creator sees own game in available list")
		}
	})
	t.Run("owner cannot join again", func(t *testing.T) {
		owner.request(t, http.MethodPost, "/game/"+waiting+"/join", nil, http.StatusConflict)
	})
	t.Run("cannot join minimax game", func(t *testing.T) {
		guest.request(t, http.MethodPost, "/game/"+robot+"/join", nil, http.StatusConflict)
	})
	guest.request(t, http.MethodPost, "/game/"+waiting+"/join", nil)
	joined := guest.gameDetails(t, waiting)
	if joined.FirstPlayerUUID != ownerUUID || joined.SecondPlayerUUID != guestUUID || joined.State != "first_player_turn" {
		t.Fatalf("unexpected joined game: %+v", joined)
	}
	if got := owner.gameDetails(t, waiting); got != joined {
		t.Fatal("players see different game states")
	}
	if containsGame(outsider.availableGames(t), waiting) {
		t.Fatal("started game is still available")
	}
	t.Run("cannot join twice", func(t *testing.T) {
		guest.request(t, http.MethodPost, "/game/"+waiting+"/join", nil, http.StatusConflict)
	})
	t.Run("third player cannot join", func(t *testing.T) {
		outsider.request(t, http.MethodPost, "/game/"+waiting+"/join", nil, http.StatusConflict)
	})
	t.Run("missing game", func(t *testing.T) {
		guest.request(t, http.MethodPost, "/game/"+missingUUID+"/join", nil, http.StatusNotFound)
	})
	t.Run("invalid game UUID", func(t *testing.T) {
		guest.request(t, http.MethodPost, "/game/invalid/join", nil, http.StatusBadRequest)
	})
	if got := owner.gameDetails(t, waiting); got != joined {
		t.Fatal("rejected joins changed game state")
	}
}

func TestGameAccessAndInvalidTurns(t *testing.T) {
	first, second, outsider := newAPIClient(t), newAPIClient(t), newAPIClient(t)
	first.register(t)
	second.register(t)
	outsider.register(t)
	id := first.createGame(t, false)
	t.Run("cannot move before opponent joins", func(t *testing.T) {
		first.request(t, http.MethodPost, "/game/"+id, turnRequest{X: 0, Y: 0}, http.StatusConflict)
	})
	second.request(t, http.MethodPost, "/game/"+id+"/join", nil)
	before := first.gameDetails(t, id)
	for _, tt := range []struct {
		name, method, path string
		api                apiClient
		body               any
		status             int
	}{
		{"outsider cannot read", http.MethodGet, "/game/" + id, outsider, nil, http.StatusForbidden},
		{"outsider cannot move", http.MethodPost, "/game/" + id, outsider, turnRequest{X: 0, Y: 0}, http.StatusForbidden},
		{"not second player's turn", http.MethodPost, "/game/" + id, second, turnRequest{X: 0, Y: 0}, http.StatusConflict},
		{"negative X", http.MethodPost, "/game/" + id, first, turnRequest{X: -1, Y: 0}, http.StatusBadRequest},
		{"large Y", http.MethodPost, "/game/" + id, first, turnRequest{X: 0, Y: 3}, http.StatusBadRequest},
		{"wrong coordinate type", http.MethodPost, "/game/" + id, first, `{"X":"0","Y":0}`, http.StatusBadRequest},
		{"malformed turn", http.MethodPost, "/game/" + id, first, `{"X":`, http.StatusBadRequest},
		{"unknown turn field", http.MethodPost, "/game/" + id, first, `{"X":0,"Y":0,"extra":1}`, http.StatusBadRequest},
		{"multiple turn objects", http.MethodPost, "/game/" + id, first, `{"X":0,"Y":0}{"X":1,"Y":1}`, http.StatusBadRequest},
		{"empty turn", http.MethodPost, "/game/" + id, first, nil, http.StatusBadRequest},
		{"read missing game", http.MethodGet, "/game/" + missingUUID, first, nil, http.StatusNotFound},
		{"move in missing game", http.MethodPost, "/game/" + missingUUID, first, turnRequest{X: 0, Y: 0}, http.StatusNotFound},
		{"read invalid UUID", http.MethodGet, "/game/invalid", first, nil, http.StatusBadRequest},
		{"move with invalid UUID", http.MethodPost, "/game/invalid", first, turnRequest{X: 0, Y: 0}, http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.api.request(t, tt.method, tt.path, tt.body, tt.status)
			if got := first.gameDetails(t, id); got != before {
				t.Fatal("rejected request changed game")
			}
		})
	}
	first.game(t, http.MethodPost, id, turnRequest{X: 0, Y: 0})
	before = first.gameDetails(t, id)
	t.Run("occupied cell", func(t *testing.T) {
		second.request(t, http.MethodPost, "/game/"+id, turnRequest{X: 0, Y: 0}, http.StatusConflict)
		if got := first.gameDetails(t, id); got != before {
			t.Fatal("occupied-cell request changed game")
		}
	})
}

func TestMultiplayerGame(t *testing.T) {
	for _, tt := range []struct {
		name      string
		moves     [][2]int
		wantState string
	}{
		{"first player wins", [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {0, 2}}, "first_player_won"},
		{"second player wins", [][2]int{{0, 0}, {1, 0}, {2, 0}, {1, 1}, {2, 2}, {1, 2}}, "second_player_won"},
		{"draw", [][2]int{{0, 0}, {1, 1}, {0, 2}, {0, 1}, {2, 1}, {1, 2}, {1, 0}, {2, 0}, {2, 2}}, "draw"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			players := [2]apiClient{newAPIClient(t), newAPIClient(t)}
			players[0].register(t)
			players[1].register(t)
			id := players[0].createGame(t, false)
			players[1].request(t, http.MethodPost, "/game/"+id+"/join", nil)
			var wantTable [3][3]int
			var game gameResponse
			for turn, move := range tt.moves {
				player := turn % 2
				game = players[player].game(t, http.MethodPost, id, turnRequest{X: move[0], Y: move[1]})
				wantTable[move[0]][move[1]] = player + 1
				if game.Table != wantTable {
					t.Fatalf("turn %d: table = %v, want %v (no automatic bot move)", turn+1, game.Table, wantTable)
				}
				wantState := "second_player_turn"
				if player == 1 {
					wantState = "first_player_turn"
				}
				if turn == len(tt.moves)-1 {
					wantState = tt.wantState
				}
				if game.State != wantState {
					t.Fatalf("turn %d: state = %q, want %q", turn+1, game.State, wantState)
				}
			}
			for _, player := range players {
				if saved := player.game(t, http.MethodGet, id, nil); saved != game {
					t.Fatal("saved game differs from final turn")
				}
				if finished := player.game(t, http.MethodPost, id, turnRequest{X: 0, Y: 0}); finished != game {
					t.Fatal("finished game changed after another move")
				}
			}
		})
	}
}

type gameDetailsResponse struct {
	gameResponse
	UUID             string `json:"uuid"`
	Minimax          bool   `json:"minimax"`
	FirstPlayerUUID  string `json:"firstPlayerUUID"`
	SecondPlayerUUID string `json:"secondPlayerUUID"`
}

func (api apiClient) createGame(t *testing.T, withRobot bool) string {
	t.Helper()
	id := strings.TrimSpace(string(api.request(t, http.MethodPost, "/game", map[string]bool{"WithRobot": withRobot})))
	assertUUID(t, id)
	return id
}

func (api apiClient) gameDetails(t *testing.T, id string) gameDetailsResponse {
	t.Helper()
	data := api.request(t, http.MethodGet, "/game/"+id, nil)
	var game gameDetailsResponse
	if err := json.Unmarshal(data, &game); err != nil {
		t.Fatal(err)
	}
	return game
}

func (api apiClient) availableGames(t *testing.T) []gameDetailsResponse {
	t.Helper()
	data := api.request(t, http.MethodGet, "/game", nil)
	var games []gameDetailsResponse
	if err := json.Unmarshal(data, &games); err != nil {
		t.Fatal(err)
	}
	if games == nil {
		t.Fatalf("game list must be a JSON array, got %s", data)
	}
	return games
}

func containsGame(games []gameDetailsResponse, id string) bool {
	for _, game := range games {
		if game.UUID == id {
			return true
		}
	}
	return false
}

type apiClient struct {
	baseURL  string
	client   *http.Client
	username string
	password string
}

type turnRequest struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

type gameResponse struct {
	Table [3][3]int `json:"table"`
	State string    `json:"state"`
}

func newAPIClient(t *testing.T) apiClient {
	t.Helper()
	baseURL := strings.TrimRight(os.Getenv("E2E_BASE_URL"), "/")
	if baseURL == "" {
		t.Skip("run task e2e from src to start the isolated test system")
	}
	client := &http.Client{
		Timeout:       10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	t.Cleanup(client.CloseIdleConnections)
	return apiClient{baseURL: baseURL, client: client, username: "e2e-" + rand.Text()[:16], password: "e2e-password"}
}

// request sends an authenticated HTTP request and checks its status.
// A string body is sent unchanged, allowing malformed JSON tests.
func (api apiClient) request(t *testing.T, method, path string, body any, expectedStatus ...int) []byte {
	t.Helper()
	wantStatus := http.StatusOK
	if len(expectedStatus) > 0 {
		wantStatus = expectedStatus[0]
	}

	var payload []byte
	if raw, ok := body.(string); ok {
		payload = []byte(raw)
	} else if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("encode request: %v", err)
		}
	}
	request, err := http.NewRequestWithContext(t.Context(), method, api.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if api.username != "" || api.password != "" {
		request.SetBasicAuth(api.username, api.password)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := api.client.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read %s %s response: %v", method, path, err)
	}
	t.Logf("%s %s -> %d", method, path, response.StatusCode)
	if response.StatusCode != wantStatus {
		t.Fatalf("%s %s: status = %d, want %d; body = %q", method, path, response.StatusCode, wantStatus, data)
	}
	if wantStatus == http.StatusUnauthorized && response.Header.Get("WWW-Authenticate") == "" {
		t.Fatalf("%s %s: missing WWW-Authenticate header", method, path)
	}
	return data
}

func (api apiClient) game(t *testing.T, method, gameUUID string, body any) gameResponse {
	t.Helper()
	data := api.request(t, method, "/game/"+gameUUID, body)
	var game gameResponse
	if err := json.Unmarshal(data, &game); err != nil {
		t.Fatalf("decode game response: %v; body = %q", err, data)
	}
	return game
}

func assertValidTurn(t *testing.T, previous, current [3][3]int) {
	t.Helper()

	playerTurns := 0
	minimaxTurns := 0
	for x := range current {
		for y := range current[x] {
			if current[x][y] < 0 || current[x][y] > 2 {
				t.Fatalf("invalid cell value %d at (%d, %d)", current[x][y], x, y)
			}
			if previous[x][y] != 0 && previous[x][y] != current[x][y] {
				t.Fatalf(
					"existing cell changed at (%d, %d): before = %d, after = %d",
					x,
					y,
					previous[x][y],
					current[x][y],
				)
			}
			if previous[x][y] == 0 {
				switch current[x][y] {
				case 1:
					playerTurns++
				case 2:
					minimaxTurns++
				}
			}
		}
	}

	if playerTurns != 1 {
		t.Fatalf("player added %d marks in one turn, want 1", playerTurns)
	}
	if minimaxTurns > 1 {
		t.Fatalf("minimax added %d marks in one turn, want at most 1", minimaxTurns)
	}
}

func gameFinished(state string) bool {
	switch state {
	case "first_player_won", "second_player_won", "draw":
		return true
	default:
		return false
	}
}
