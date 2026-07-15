package web

import (
	"context"
	"datasource"
	"domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"go.uber.org/fx"
)

func InitHTTPServer(lc fx.Lifecycle, db datasource.Database) *http.Server {
	mux := http.NewServeMux()
	uuidRegexp := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

	mux.HandleFunc("POST /game/{UUID}", gamePostHandler(db, *uuidRegexp))

	srv := &http.Server{
		Addr:    ":5005",
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != http.ErrServerClosed {
					return
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("SHUTDOWN")
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

func gamePostHandler(db datasource.Database, uuidRegexp regexp.Regexp) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		UUID := r.PathValue("UUID")

		var req GameRequest
		var resp GameResponse

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateRequest(req, UUID, uuidRegexp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ret, err := domain.HandleTurn(req.toDomain(), db, UUID)
		if err != nil {
			http.Error(w, "Error handling turn: "+err.Error(), http.StatusBadRequest)
			return
		}

		resp.UUID = ret.UUID
		resp.Table = ret.Table

		resp.Result = gameResult(int(ret.Status))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func gameResult(status int) string {
	switch domain.MinimaxStatus(status) {
	case domain.Draw:
		return "draw"
	case domain.MinimaxLose:
		return "player_won"
	case domain.MinimaxWin:
		return "computer_won"
	default:
		return "in_progress"
	}
}

func validateRequest(req GameRequest, UUID string, uuidRegexp regexp.Regexp) error {
	if !uuidRegexp.MatchString(UUID) {
		return errors.New("invalid uuid")
	}

	if req.Y < 0 || req.X < 0 || req.Y > 2 || req.X > 2 {
		return errors.New("invalid cordinates")
	}

	return nil
}
