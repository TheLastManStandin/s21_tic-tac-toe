package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"go.uber.org/fx"
)

func InitHTTPServer(lc fx.Lifecycle, h *Handler) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/hc", h.heathCheck)
	mux.HandleFunc("/game/{UUID}", h.handleGamePost)

	srv := &http.Server{
		Addr:    ":5005",
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("WELCOME TO THE GAME, TRAVELER. TRY localhost" + srv.Addr)
			fmt.Println("YOUR POST MUST BE /GAME/UUID")
			go func() {
				if err := srv.ListenAndServe(); err != http.ErrServerClosed {
					log.Fatalf("ListenAndServe(): %s", err)
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

func (h Handler) heathCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK\n"))
}

func (h Handler) handleGamePost(w http.ResponseWriter, r *http.Request) {
	UUID := r.PathValue("UUID")

	var req GameRequest
	if err := validateUUID(UUID, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.service.HandleTurn(req.ToDomainGame(UUID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(formatBoard(ret.Table)))
}

func validateUUID(UUID string, req *GameRequest) error {
	var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(UUID) {
		return errors.New("invalid UUID format")
	}
	//req.UUID = UUID
	return nil
}

func formatBoard(table [3][3]int8) string {
	var b strings.Builder

	b.WriteByte('\n')
	for i, row := range table {
		fmt.Fprintf(&b, " %d | %d | %d ", row[0], row[1], row[2])
		if i < len(table)-1 {
			b.WriteString("\n-----------\n")
		}
	}

	b.WriteByte('\n')
	b.WriteByte('\n')
	return b.String()
}
