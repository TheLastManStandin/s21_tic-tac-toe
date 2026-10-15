package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"tickTackToe/internal/domain/services"
	"time"

	"tickTackToe/internal/datasource"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/repositorys"
	"tickTackToe/internal/migrations"
	"tickTackToe/internal/web/handlers"
	"tickTackToe/internal/web/middleware"
	"tickTackToe/internal/web/router"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func TestGameAgainstMinimax(t *testing.T) {
	tests := []struct {
		name      string
		gameUUID  string
		moves     [][2]int
		wantState string
	}{
		{
			name:      "minimax wins",
			gameUUID:  "00000000-0000-0000-0000-000000000001",
			moves:     [][2]int{{0, 0}, {0, 1}, {1, 0}},
			wantState: domain.SecondPlayerWon.String(),
		},
		{
			name:      "draw",
			gameUUID:  "00000000-0000-0000-0000-000000000002",
			moves:     [][2]int{{0, 0}, {0, 1}, {2, 0}, {1, 2}, {2, 2}},
			wantState: domain.Draw.String(),
		},
	}

	gameUUIDs := make([]string, 0, len(tests))
	for _, tt := range tests {
		gameUUIDs = append(gameUUIDs, tt.gameUUID)
	}

	server, repository := startTestApplication(t, gameUUIDs)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var game gameResponse
			for turn, move := range tt.moves {
				x, y := move[0], move[1]
				if game.Table[x][y] != 0 {
					t.Fatalf("turn %d targets occupied cell (%d, %d): table = %v", turn+1, x, y, game.Table)
				}

				previousTable := game.Table
				game = makeTurn(t, server.Client(), server.URL, tt.gameUUID, x, y)
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

			assertGameSaved(t, repository, tt.gameUUID, game)

			finishedGame := makeTurn(t, server.Client(), server.URL, tt.gameUUID, 0, 0)
			if finishedGame != game {
				t.Fatalf(
					"finished game changed after another turn: before = %+v, after = %+v",
					game,
					finishedGame,
				)
			}
		})
	}
}

func startTestApplication(
	t *testing.T,
	gameUUIDs []string,
) (*httptest.Server, repositorys.GameRepository) {
	t.Helper()
	const (
		testUsername = "e2e-player"
		testPassword = "e2e-password"
	)

	if os.Getenv("DATABASE_CONNECTION_STRING") == "" {
		_ = godotenv.Load("../.env")
	}
	connectionString := os.Getenv("DATABASE_CONNECTION_STRING")
	if connectionString == "" {
		t.Skip("PostgreSQL e2e tests require DATABASE_CONNECTION_STRING")
	}

	connectionContext, cancelConnection := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelConnection()

	database, err := pgxpool.New(connectionContext, connectionString)
	if err != nil {
		t.Skipf("PostgreSQL is not available: %v", err)
	}
	if err := database.Ping(connectionContext); err != nil {
		database.Close()
		t.Skipf("PostgreSQL is not available: %v", err)
	}

	if err := migrations.Up(); err != nil {
		database.Close()
		t.Fatalf("apply database migrations: %v", err)
	}

	deleteTestGames := func(ctx context.Context) {
		for _, gameUUID := range gameUUIDs {
			if _, err := database.Exec(ctx, "DELETE FROM games WHERE uuid = $1", gameUUID); err != nil {
				t.Errorf("delete test game %s: %v", gameUUID, err)
			}
		}
	}
	deleteTestGames(connectionContext)
	if _, err := database.Exec(
		connectionContext,
		"DELETE FROM client WHERE name = $1",
		testUsername,
	); err != nil {
		database.Close()
		t.Fatalf("delete test user: %v", err)
	}

	var mux *http.ServeMux
	var repository repositorys.GameRepository
	var userService *services.UserService
	application := fx.New(
		fx.NopLogger,
		fx.Provide(
			zap.NewNop,
			router.NewServeMux,
			handlers.NewGameHandler,
			handlers.NewAuthHandler,
			handlers.NewUserHandler,
			services.NewGameService,
			services.NewUserService,
			middleware.NewBasicAuth,
			func(service *services.UserService) middleware.UserAuthenticator {
				return service
			},
			fx.Annotate(
				datasource.NewPostgresDatabase,
				fx.As(new(repositorys.GameRepository)),
				fx.As(new(repositorys.UserRepository)),
			),
		),
		fx.Populate(&mux, &repository, &userService),
	)

	startContext, cancelStart := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelStart()
	if err := application.Start(startContext); err != nil {
		database.Close()
		t.Fatalf("start test application: %v", err)
	}

	userUUID, err := userService.RegisterUser(
		context.Background(),
		testUsername,
		testPassword,
	)
	if err != nil {
		_ = application.Stop(context.Background())
		database.Close()
		t.Fatalf("register test user: %v", err)
	}
	for _, gameUUID := range gameUUIDs {
		if err := repository.InitGame(
			context.Background(),
			gameUUID,
			userUUID,
			true,
		); err != nil {
			_ = application.Stop(context.Background())
			database.Close()
			t.Fatalf("initialize test game %s: %v", gameUUID, err)
		}
	}

	server := httptest.NewServer(mux)
	t.Cleanup(func() {
		server.Close()

		stopContext, cancelStop := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelStop()
		if err := application.Stop(stopContext); err != nil {
			t.Errorf("stop test application: %v", err)
		}

		deleteTestGames(stopContext)
		if _, err := database.Exec(
			stopContext,
			"DELETE FROM client WHERE name = $1",
			testUsername,
		); err != nil {
			t.Errorf("delete test user: %v", err)
		}
		database.Close()
	})

	return server, repository
}

type turnRequest struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

type gameResponse struct {
	Table [3][3]int `json:"table"`
	State string    `json:"state"`
}

func makeTurn(
	t *testing.T,
	client *http.Client,
	serverURL string,
	gameUUID string,
	x, y int,
) gameResponse {
	t.Helper()

	body, err := json.Marshal(turnRequest{X: x, Y: y})
	if err != nil {
		t.Fatalf("encode turn request: %v", err)
	}

	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		fmt.Sprintf("%s/game/%s", serverURL, gameUUID),
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create turn request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth("e2e-player", "e2e-password")

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("make turn request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		var errorResponse bytes.Buffer
		_, _ = errorResponse.ReadFrom(response.Body)
		t.Fatalf("make turn status = %d, body = %q", response.StatusCode, errorResponse.String())
	}

	var game gameResponse
	if err := json.NewDecoder(response.Body).Decode(&game); err != nil {
		t.Fatalf("decode turn response: %v", err)
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

func assertGameSaved(
	t *testing.T,
	repository repositorys.GameRepository,
	gameUUID string,
	want gameResponse,
) {
	t.Helper()

	savedGame, err := repository.FindGameByID(context.Background(), gameUUID)
	if err != nil {
		t.Fatalf("load saved game: %v", err)
	}
	if savedGame.Table != want.Table {
		t.Fatalf("saved table = %v, response table = %v", savedGame.Table, want.Table)
	}
	if savedGame.State.String() != want.State {
		t.Fatalf("saved state = %q, response state = %q", savedGame.State.String(), want.State)
	}
}

func gameFinished(state string) bool {
	switch state {
	case domain.FirstPlayerWon.String(), domain.SecondPlayerWon.String(), domain.Draw.String():
		return true
	default:
		return false
	}
}
